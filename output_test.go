package main

import (
	"strings"
	"testing"
)

func TestParseOutputNonEmpty(t *testing.T) {

	reportText := "\n```\nHere is a terraform plan for example\n```\n\n# Outputs:\n```json " + outputsStartIndicator + "\n" + `
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

	parsedOutputs := parseOutputsFromText(&reportText)
	if strings.TrimSpace(expectedOutputs) != parsedOutputs {
		t.Error()
	}

}

func TestParseOutputEmpty(t *testing.T) {

	reportText := "\n```\nHere is a terraform plan for example\n```"

	parsedOutputs := parseOutputsFromText(&reportText)
	if "" != parsedOutputs {
		t.Error()
	}

}
