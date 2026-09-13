package utils

import (
	"os"
	"regexp"
	"unicode/utf8"

	"charm.land/bubbles/v2/table"
)

func GetToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func GetHeaderWidth(t *table.Model) int {
	w := 0
	for _, col := range t.Columns() {
		w += col.Width
	}
	return w
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func LenWithoutANSI(s string) int {
	clean := ansiRegexp.ReplaceAllString(s, "")
	return utf8.RuneCountInString(clean)
}
