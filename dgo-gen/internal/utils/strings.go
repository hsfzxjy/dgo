package utils

import (
	"regexp"
	"strings"
)

var reComma = regexp.MustCompile(`\s*,\s*`)

func ParseDirectives(spec string) []string {
	trailing := strings.TrimSpace(spec)
	if len(trailing) == 0 {
		return nil
	}
	return reComma.Split(trailing, -1)
}
