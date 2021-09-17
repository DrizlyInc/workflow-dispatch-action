package main

import (
	"strings"
)

const outputsStartIndicator = "---- BEGIN CHECK OUTPUT ----"

// parseOutputsFromText takes a string and parses out any "outputs"
// included therein. The outputs are delimited using a code block (```)
// annotated with the outputsStartIndicator
func parseOutputsFromText(reportText *string) string {
	if reportText == nil {
		return ""
	}

	splitReport := strings.Split(*reportText, outputsStartIndicator)
	if len(splitReport) < 2 {
		return ""
	}

	outputs := strings.Split(splitReport[1], "```")[0]
	return strings.TrimSpace(outputs)
}
