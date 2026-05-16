package providers

import (
	"fmt"
	"strconv"
)

// deref returns *s or "" when s is nil. Shared across all providers.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// joinExpiry stringifies a month/year tuple as "MM/YYYY", matching how the TS
// providers format card expiry on the meta envelope. Both inputs may be
// numeric or string at the JSON layer.
func joinExpiry(month, year any) string {
	return fmt.Sprintf("%s/%s", strOrXX(month, 2), strOrXX(year, 4))
}

func strOrXX(v any, width int) string {
	if v == nil {
		return repeat("X", width)
	}
	switch x := v.(type) {
	case string:
		if x == "" {
			return repeat("X", width)
		}
		return x
	case float64:
		return strconv.Itoa(int(x))
	case int:
		return strconv.Itoa(x)
	}
	return repeat("X", width)
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// numericString returns v as a string regardless of its JSON-decoded type
// (numbers come back as float64).
func numericString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	}
	return ""
}
