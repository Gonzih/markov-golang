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

	"github.com/julienschmidt/httprouter"
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

func init() {
	flag.IntVar(&maxNumberOfSentences, "sentences", 100, "number of sentences to generate")
	flag.Parse()
	rand.Seed(time.Now().Unix())
	sentenceEndRegexp = regexp.MustCompile("^.*[.!?]$")
	input := ReadInput()
	sharedChain = GenerateChain(input)
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

func main() {
	router := httprouter.New()
	router.GET("/", TalkHandler)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf(":%s", port)

	log.Fatal(http.ListenAndServe(address, router))
}

type TemplatePayload struct {
	Output string
}

func RenderTemplate(payload TemplatePayload, w http.ResponseWriter) error {
	tpl, err := template.ParseFiles("templates/index.html")

	if err != nil {
		log.Fatalf("Error parsing template: %s", err.Error())
		return err
	}

	return tpl.ExecuteTemplate(w, "index.html", payload)
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

	err = RenderTemplate(TemplatePayload{Output: output}, w)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// fmt.Fprint(w, fmt.Sprintf("%s\n", output))
}
