package text

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
)

func Cleanup(text string) (string, error) {
	// windows new line, Github, really ?
	text = strings.Replace(text, "\r\n", "\n", -1)

	// remove all unicode control characters except
	// '\n', '\r' and '\t'
	t := runes.Remove(runes.Predicate(func(r rune) bool {
		switch r {
		case '\r', '\n', '\t':
			return false
		}
		return unicode.IsControl(r)
	}))
	sanitized, _, err := transform.String(t, text)
	if err != nil {
		return "", err
	}

	// trim extra new line not displayed in the github UI but still present in the data
	return strings.TrimSpace(sanitized), nil
}
