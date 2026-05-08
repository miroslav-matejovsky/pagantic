package constraint

import "strings"

// RepairJSON closes common truncated JSON output.
//
// It trims trailing space, tracks strings and open containers, and appends
// missing closing quotes, braces, or brackets. It is not general JSON repair.
// Use json.Valid after repair.
func RepairJSON(s string) string {
	s = strings.TrimRight(s, " \t\n\r")
	if s == "" {
		return s
	}

	var stack []byte
	inString := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '{':
			stack = append(stack, '}')
		case '[':
			stack = append(stack, ']')
		case '}', ']':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	if inString {
		s += `"`
	}
	sep := ""
	if strings.Contains(s, "\n") {
		sep = "\n"
	}
	for i := len(stack) - 1; i >= 0; i-- {
		s += sep + string(stack[i])
	}
	return s
}
