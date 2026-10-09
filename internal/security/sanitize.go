package security

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	ansiRegex = regexp.MustCompile(`\x1b(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~]|\][^\x07\x1b]*(\x07|\x1b\\))`)
	nickRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{2,16}$`)
)

func StripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

func StripControlChars(str string, allowNewlines bool) string {
	var builder strings.Builder
	builder.Grow(len(str))

	for _, r := range str {
		if r == '\n' && allowNewlines {
			builder.WriteRune(r)
			continue
		}
		if unicode.IsControl(r) {
			continue
		}
		if r == 0x07 || r == 0x08 || r == 0x1b || r == 0x7f {
			continue
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

func SanitizeText(input string, maxLen int, allowNewlines bool) string {
	clean := StripANSI(input)
	clean = StripControlChars(clean, allowNewlines)
	clean = strings.TrimSpace(clean)

	runes := []rune(clean)
	if maxLen > 0 && len(runes) > maxLen {
		clean = string(runes[:maxLen])
	}
	return clean
}

func ValidateNickname(nick string) (string, error) {
	clean := StripANSI(nick)
	clean = StripControlChars(clean, false)
	clean = strings.TrimSpace(clean)

	if len(clean) < 2 {
		return "", errors.New("nickname must be at least 2 characters")
	}
	if len(clean) > 16 {
		return "", errors.New("nickname cannot exceed 16 characters")
	}
	if !nickRegex.MatchString(clean) {
		return "", errors.New("nickname may only contain letters, numbers, '-', '_', '.'")
	}

	lower := strings.ToLower(clean)
	reserved := []string{"system", "admin", "root", "server", "afterdark", "anonymous"}
	for _, r := range reserved {
		if lower == r {
			return "", fmt.Errorf("nickname '%s' is reserved", clean)
		}
	}

	return clean, nil
}
