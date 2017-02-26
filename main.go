package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Chain map[string][]string

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadInput() string {
	file, err := os.Open("input.csv")
	checkErr(err)
	defer file.Close()

	reader := bufio.NewReader(file)

	var input string

	csvReader := csv.NewReader(reader)

	records, err := csvReader.ReadAll()

	checkErr(err)

	for _, record := range records {
		if len(record) > 0 {
			input = input + record[0] + ". "
		}
	}

	return input
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

	// fmt.Printf("%v\n", chain)

	return chain
}

func GenerateOutput(chain Chain) string {
	var output string

	current := "This"

	rand.Seed(time.Now().UnixNano())

	for {
		nextArr := chain[current]
		l := len(nextArr)

		if l == 0 {
			break
		}

		i := rand.Intn(len(nextArr))
		next := nextArr[i]
		output = output + " " + current
		current = next

		if strings.HasSuffix(next, ".") {
			output = output + " " + next
			break
		}
	}

	output = strings.Trim(output, " ")

	return output
}

type Message struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	Icon     string `json:"icon_emoji"`
}

func PostToSlack(output string) {
	message := Message{Text: output, Username: "Markov", Icon: ":shipit:"}
	json, err := json.Marshal(&message)

	checkErr(err)

	url := "https://hooks.slack.com/services/000000000/000000000/000000000000000000000000"

	client := http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))

	checkErr(err)

	req.Header.Set("Content-Type", "application/json")
	_, err = client.Do(req)

	checkErr(err)

	req.Body.Close()
}

func main() {
	input := ReadInput()
	chain := GenerateChain(input)
	output := GenerateOutput(chain)

	fmt.Printf("%s\n", output)

	PostToSlack(output)
}
