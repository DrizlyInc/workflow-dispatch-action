package main

import (
	"strings"
)

func parseOutputsFromText(reportText *string) string {

	splitReport := strings.Split(*reportText, "---- BEGIN CHECK OUTPUT ----")
	if len(splitReport) < 2 {
		return ""
	}

	outputs := strings.Split(splitReport[1], "```")[0]
	return strings.TrimSpace(outputs)
}
