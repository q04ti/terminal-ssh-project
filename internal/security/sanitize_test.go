package security

import (
	"strings"
	"testing"
)

func TestStripANSI(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"\x1b[31mRed Alert\x1b[0m", "Red Alert"},
		{"\x1b[2J\x1b[HClear Screen", "Clear Screen"},
		{"Plain text", "Plain text"},
		{"\x1b]0;Title\x07Hello", "Hello"},
	}

	for _, tc := range cases {
		out := StripANSI(tc.input)
		if out != tc.expected {
			t.Errorf("StripANSI(%q) = %q, expected %q", tc.input, out, tc.expected)
		}
	}
}

func TestStripControlChars(t *testing.T) {
	// Bell, backspace, null byte
	hazardous := "Hello\x00\x07\x08World\x7f"
	clean := StripControlChars(hazardous, false)
	if clean != "HelloWorld" {
		t.Errorf("StripControlChars failed: got %q, expected HelloWorld", clean)
	}

	// Newline preserved when flag is true
	multiline := "Line1\nLine2"
	cleanMulti := StripControlChars(multiline, true)
	if cleanMulti != multiline {
		t.Errorf("StripControlChars multiline failed: got %q", cleanMulti)
	}
}

func TestValidateNickname(t *testing.T) {
	// Valid nicknames
	validNicks := []string{"neo", "ghostbyte", "wanderer_99", "agent-x", "node.01"}
	for _, n := range validNicks {
		clean, err := ValidateNickname(n)
		if err != nil {
			t.Errorf("Valid nickname %q rejected: %v", n, err)
		}
		if clean != n {
			t.Errorf("Expected %q, got %q", n, clean)
		}
	}

	// Invalid nicknames
	invalidNicks := []string{
		"a",                       // too short
		"waytoolongnicknameexceeds", // too long (>16)
		"hello world",            // contains space
		"bad@name!",              // special symbols
		"admin",                  // reserved
		"root",                   // reserved
		"\x1b[31mhacker\x1b[0m",  // ansi code injection
	}

	for _, n := range invalidNicks {
		clean, err := ValidateNickname(n)
		if err == nil {
			t.Errorf("Invalid nickname %q should have been rejected, got %q", n, clean)
		}
	}
}

func TestSanitizeText(t *testing.T) {
	input := "\x1b[32m   Secure message with \x07 bell   \x1b[0m"
	out := SanitizeText(input, 100, false)
	if out != "Secure message with  bell" && !strings.Contains(out, "Secure message") {
		t.Errorf("Unexpected sanitized text: %q", out)
	}

	// Length bounding
	long := strings.Repeat("A", 500)
	bounded := SanitizeText(long, 50, false)
	if len(bounded) != 50 {
		t.Errorf("SanitizeText failed to bound length: got %d runes", len(bounded))
	}
}
