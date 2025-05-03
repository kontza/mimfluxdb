package cmd

import (
	"strings"
	"unicode"
)

func toTitleCase(s string) string {
	var titleCase strings.Builder
	words := strings.Fields(s)
	for i, word := range words {
		if i > 0 {
			titleCase.WriteRune(' ')
		}
		titleCase.WriteRune(unicode.ToUpper(rune(word[0])))
		titleCase.WriteString(word[1:])
	}
	return titleCase.String()
}
