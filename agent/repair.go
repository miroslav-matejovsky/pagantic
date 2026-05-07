package agent

import "strings"

// repairJSON attempts to complete truncated JSON by appending missing closing
// characters. This is a targeted fallback for a grammar-constrained LLM
// generation quirk where the engine stops before emitting the final `}` or `]`.
//
// It is NOT a general JSON fixer -- it only closes unclosed containers and
// strings. Callers must validate the result with json.Valid after repair.
//
// If the input is already valid or not JSON-like, it is returned trimmed.
func repairJSON(s string) string {
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
