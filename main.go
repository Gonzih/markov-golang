package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
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

func main() {
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

	chain := make(map[string][]string)

	words := strings.Fields(input)

	for i, word := range words {
		trimmed := strings.Trim(word, " \t\n")
		dict := chain[trimmed]

		if i < len(words)-1 && !contains(dict, trimmed) {
			chain[trimmed] = append(dict, words[i+1])
		}
	}

	// fmt.Printf("%v\n", chain)

	var output string

	current := "As"

	rand.Seed(time.Now().Unix())

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

	fmt.Printf("\"%s\"\n", output)
}
