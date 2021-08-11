package main

import (
	"strings"
	"testing"
)

func TestParseOutputNonEmpty(t *testing.T) {

	reportText := "\n```\nHere is a terraform plan for example\n```\n\n# Outputs:\n```json ---- BEGIN CHECK OUTPUT ----\n" + `
{
	"my_output": "my_value"
}
` + "```\n"

	expectedOutputs :=
		`
{
	"my_output": "my_value"
}
`

	parsedOutputs := parseOutputs(&reportText)
	if strings.TrimSpace(expectedOutputs) != parsedOutputs {
		t.Error()
	}

}

func TestParseOutputEmpty(t *testing.T) {

	reportText := "\n```\nHere is a terraform plan for example\n```"

	parsedOutputs := parseOutputs(&reportText)
	if "" != parsedOutputs {
		t.Error()
	}

}
