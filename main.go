package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
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

	// fmt.Printf("%v\n", chain)

	return chain
}

func GenerateSentence(chain *Chain) string {
	var output string

	current := "It"

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

		if strings.HasSuffix(next, ".") {
			output = output + " " + next
			break
		}
	}

	output = strings.Trim(output, " ")

	return output
}

func GenerateOutput(chain Chain) string {
	return GenerateSentence(&chain)
}

func main() {
	input := ReadInput()
	chain := GenerateChain(input)
	output := GenerateOutput(chain)

	fmt.Printf("%s\n", output)
}
