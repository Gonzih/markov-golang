package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Chain map[string][]string

var maxNumberOfSentences int
var sharedChain Chain
var redisClient *redis.Client
var templates *template.Template
var sentenceEndRegexp = regexp.MustCompile("^.*[.!?]$")

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadInput() string {
	file, err := os.Open("input.txt")
	checkErr(err)
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, file)

	return string(buf.Bytes())
}

func GenerateChain(input string) Chain {
	chain := make(Chain)

	words := strings.Fields(input)

	for i, word := range words {
		trimmed := strings.Trim(word, " \t\n")
		dict := chain[trimmed]

		if i < len(words)-1 {
			next := words[i+1]

			chain[trimmed] = append(dict, next)
		}
	}

	// log.Printf("%v\n", chain)

	return chain
}

func GenerateSentence(current string, chain *Chain, limit int) string {
	var output string
	var count int

	rand.Seed(time.Now().UnixNano())

	for {
		nextArr := (*chain)[current]
		l := len(nextArr)

		if l == 0 {
			break
		}

		i := rand.Intn(len(nextArr))
		next := nextArr[i]
		output = output + " " + current
		current = next

		if sentenceEndRegexp.MatchString(next) {
			count++

			if count >= limit {
				output = output + " " + next
				break
			}
		}
	}

	output = strings.Trim(output, " ")

	return output
}

func RandomBeginningOfASentence(chain *Chain) string {
	for {
		var i int
		n := rand.Intn(len(*chain))

		for m := range *chain {
			b := []byte(m[0:1])
			r, _ := utf8.DecodeRune(b)
			isUpper := unicode.IsUpper(r)
			if i == n && isUpper {
				return m
			}
			i++
		}
	}

	return "test"
}

func GenerateQuote(chain *Chain, limit int) string {
	start := RandomBeginningOfASentence(chain)
	return GenerateSentence(start, chain, limit)
}

func generate(n int) string {
	output := GenerateQuote(&sharedChain, n)
	log.Printf("%s\n", output)

	return output
}

func newRedisClient() *redis.Client {
	opts, err := redis.ParseURL(viper.GetString("redis_url"))

	if err != nil {
		log.Fatalf("Error initializing redis client: %s", err.Error())
	}

	client := redis.NewClient(opts)

	_, err = client.Ping().Result()

	if err != nil {
		log.Fatalf("Error executing pong on redis, %s", err.Error())
	}

	return client
}

func initViper() {
	pflag.Int("sentences", 100, "number of sentences to generate")
	pflag.Parse()

	viper.AddConfigPath("./config")
	viper.SetConfigName("development")
	viper.SetConfigType("yaml")
	viper.BindEnv("redis_url")
	viper.BindEnv("environment")
	viper.BindPFlags(pflag.CommandLine)

	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error reading viper conf: %s", err.Error())
	}
}

func init() {
	initViper()

	rand.Seed(time.Now().Unix())
	templates = template.Must(template.ParseGlob("templates/*"))
	maxNumberOfSentences = viper.GetInt("sentences")
	input := ReadInput()
	sharedChain = GenerateChain(input)
	redisClient = newRedisClient()
}

func main() {
	port := os.Getenv("PORT")

	router := httprouter.New()
	router.GET("/", IndexHandler)
	router.GET("/talk/:id", ShowHandler)

	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf(":%s", port)

	log.Fatal(http.ListenAndServe(address, router))
}

type templatePayload struct {
	Quote string
	ID    string
}

func renderTemplate(payload templatePayload, w http.ResponseWriter) error {
	if viper.GetString("environment") != "production" {
		templates = template.Must(template.ParseGlob("templates/*"))
	}

	return templates.ExecuteTemplate(w, "index.html", payload)
}

func saveQuote(key, quote string) error {
	return redisClient.Set(key, quote, 0).Err()
}

func loadQuote(key string) (string, error) {
	return redisClient.Get(key).Result()
}

func ShowHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	quote, err := loadQuote(id)

	if err != nil {
		log.Printf("Error loading quote to redis: %s", err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if len(quote) == 0 {
		log.Printf("Quote not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = renderTemplate(templatePayload{Quote: quote, ID: id}, w)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var sentences int

	ns := r.FormValue("sentences")
	n, err := strconv.ParseInt(ns, 10, 64)

	if err != nil {
		sentences = 10
	} else {
		sentences = int(n)
	}

	if sentences > maxNumberOfSentences {
		fmt.Fprint(w, "Too many sentences\n")
		return
	}

	output := generate(sentences)
	id := uuid.New().String()

	err = saveQuote(id, output)

	if err != nil {
		log.Printf("Error saving quote to redis: %s", err.Error())
	}

	err = renderTemplate(templatePayload{Quote: output, ID: id}, w)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
}
