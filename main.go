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

func contains(arr []string, s string) bool {
	for _, ss := range arr {
		if ss == s {
			return true
		}
	}

	return false
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func readInput() string {
	file, err := os.Open("input.csv")
	checkErr(err)

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

func generateChain(input string) map[string][]string {
	chain := make(map[string][]string)

	words := strings.Fields(input)

	for i, word := range words {
		trimmed := strings.Trim(word, " \t\n")
		dict := chain[trimmed]

		if i < len(words)-1 {
			next := words[i+1]

			if !contains(dict, next) {
				// fmt.Printf("Adding '%s' to %v\n", next, dict)
				chain[trimmed] = append(dict, next)
			}
		}
	}

	// fmt.Printf("%v\n", chain)

	return chain
}

func generateOutput(chain map[string][]string) string {
	var output string

	current := "As"

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

		if strings.Contains(next, ".") {
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

func postToSlack(output string) {
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
	input := readInput()
	chain := generateChain(input)
	output := generateOutput(chain)

	fmt.Printf("%s\n", output)

	postToSlack(output)
}
