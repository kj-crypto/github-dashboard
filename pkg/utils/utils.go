package utils

import (
	"os"
	"regexp"
	"unicode/utf8"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func GetToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func GetHeaderWidth(t *table.Model, s *lipgloss.Style) int {
	w := 0
	p := s.GetHorizontalFrameSize()
	for _, col := range t.Columns() {
		w += col.Width + p
	}
	return w
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func LenWithoutANSI(s string) int {
	clean := ansiRegexp.ReplaceAllString(s, "")
	return utf8.RuneCountInString(clean)
}
