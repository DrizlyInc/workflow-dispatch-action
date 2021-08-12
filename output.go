package main

import (
	"strings"
)

const outputsStartIndicator = "---- BEGIN CHECK OUTPUT ----"

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
