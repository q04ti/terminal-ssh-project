package security

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// ansiRegex matches ANSI escape sequences: CSI, OSC, and 2-byte escape sequences
	ansiRegex = regexp.MustCompile(`\x1b\[[0-9:;<=>?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)|\x1b[@-Z\\-_]`)

	// nickRegex matches allowed nickname characters: letters, numbers, underscores, dashes
	nickRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{2,16}$`)
)

// StripANSI removes all ANSI escape sequences from input string.
func StripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// StripControlChars removes non-printable ASCII / control characters.
// If allowNewlines is true, standard '\n' is kept.
func StripControlChars(str string, allowNewlines bool) string {
	var builder strings.Builder
	builder.Grow(len(str))

	for _, r := range str {
		if r == '\n' && allowNewlines {
			builder.WriteRune(r)
			continue
		}
		// Strip control characters (C0 and C1 sets, DEL)
		if unicode.IsControl(r) {
			continue
		}
		// Also filter specific terminal hazard runes
		if r == 0x07 || r == 0x08 || r == 0x1b || r == 0x7f {
			continue
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

// SanitizeText removes ANSI escape codes and control characters, and trims whitespace.
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

// ValidateNickname checks if a nickname is valid.
func ValidateNickname(nick string) (string, error) {
	// Reject if raw input contains ANSI escape codes or control characters
	if StripANSI(nick) != nick || StripControlChars(nick, false) != nick {
		return "", errors.New("nickname cannot contain escape sequences or control characters")
	}

	clean := strings.TrimSpace(nick)

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
	reserved := []string{"system", "admin", "root", "server", "afterdark", "anonymous", "operator", "daemon"}
	for _, r := range reserved {
		if lower == r {
			return "", fmt.Errorf("nickname '%s' is reserved", clean)
		}
	}

	return clean, nil
}
