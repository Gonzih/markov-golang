package main

import (
	"bytes"
	"flag"
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
	"github.com/spf13/viper"
)

type Chain map[string][]string

var maxNumberOfSentences int
var sharedChain Chain
var sentenceEndRegexp *regexp.Regexp

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

func GenerateSentence(current string, chain *Chain) string {
	var output string

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
			output = output + " " + next
			break
		}
	}

	output = strings.Trim(output, " ")

	return output
}

func RandomKey(chain *Chain) string {
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

func GenerateOutput(chain *Chain) string {
	start := RandomKey(chain)
	return GenerateSentence(start, chain)
}

func generate(n int) string {
	var output string
	var i int

	for i < n {
		output = output + GenerateOutput(&sharedChain) + " "
		i++
	}

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
	viper.AddConfigPath("./config")
	viper.SetConfigName("development")
	viper.SetConfigType("yaml")
	viper.BindEnv("redis_url")

	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error reading viper conf: %s", err.Error())
	}
}

var redisClient *redis.Client

func init() {
	flag.IntVar(&maxNumberOfSentences, "sentences", 100, "number of sentences to generate")
	flag.Parse()
	rand.Seed(time.Now().Unix())
	sentenceEndRegexp = regexp.MustCompile("^.*[.!?]$")
	input := ReadInput()
	sharedChain = GenerateChain(input)
	initViper()
	redisClient = newRedisClient()
}

func main() {
	port := os.Getenv("PORT")

	router := httprouter.New()
	router.GET("/", TalkHandler)

	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf(":%s", port)

	log.Fatal(http.ListenAndServe(address, router))
}

type templatePayload struct {
	Output string
	ID     string
}

var templates = template.Must(template.ParseGlob("templates/*"))

func renderTemplate(payload templatePayload, w http.ResponseWriter) error {
	// development
	templates = template.Must(template.ParseGlob("templates/*"))

	return templates.ExecuteTemplate(w, "index.html", payload)
}

func saveQuote(key, quote string) error {
	return redisClient.Set(key, quote, 0).Err()
}

func loadQuote(key string) (string, error) {
	return redisClient.Get(key).Result()
}

func TalkHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	err = renderTemplate(templatePayload{Output: output, ID: id}, w)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
}
