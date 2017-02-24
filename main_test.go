package main

import (
	"testing"
)

func TestGenerateChain(t *testing.T) {
	chain := GenerateChain("As a user I want.")

	expects := map[string]string{
		"I":  "want.",
		"As": "a",
		"a":  "user",
	}

	for k, v := range expects {
		check := chain[k][0]
		if check != v {
			t.Errorf("Check for key '%s' failed, '%s' != '%s'. \n Chain was '%v' \n", k, check, v, chain)
		}
	}
}

func TestGenerateOutput(t *testing.T) {
	input := "This is super great thing."
	chain := GenerateChain(input)
	output := GenerateOutput(chain)

	if input != output {
		t.Errorf("%s != %s", input, output)
	}
}
