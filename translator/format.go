package translator

import (
	"fmt"
	"strings"
	"unicode"
)

func formatID(words ...string) string {
	formatted := make([]string, 0, len(words))

	for word := range partsChan(words...) {
		formatted = append(formatted, title(word))
	}

	return strings.Join(formatted, "")
}

func formatVar(words ...string) string {
	formatted := make([]string, 0, len(words))

	first := true
	for word := range partsChan(words...) {
		if first {
			word = strings.ToLower(word)
			first = false
		} else {
			word = title(word)
		}
		formatted = append(formatted, word)
	}

	return strings.Join(formatted, "")
}

func partsChan(words ...string) <-chan string {
	c := make(chan string)
	go func() {
		for _, word := range words {
			w := ""
			var last rune
			for _, ch := range word {
				if ch == '-' || ch == '_' || unicode.IsSpace(ch) {
					c <- w
					w = ""
					continue
				}
				if unicode.IsUpper(ch) && len(w) > 0 && !unicode.IsUpper(last) {
					c <- w
					w = ""
				}
				w += string(ch)
				last = ch
			}
			if w != "" {
				c <- w
			}
		}

		close(c)
	}()

	return c
}

var reserved = []string{
	"break", "default", "func", "interface", "select", "case", "defer", "go",
	"map", "struct", "chan", "else", "goto", "package", "switch", "const",
	"fallthrough", "if", "range", "type", "continue", "for", "import",
	"return", "var", "string", "bool", "int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64",
	"complex64", "complex128", "byte", "rune", "uintptr",
}

func formatReserved(s, c string) string {
	for _, r := range reserved {
		if s == r {
			return fmt.Sprintf("%s%s", strings.ToLower(c), strings.Title(s))
		}
	}

	return s
}

var acronyms = []string{
	"ID", "UUID", "JSON", "XML", "ACL", "URL", "SSO",
}

func title(word string) string {
	for _, ac := range acronyms {
		if strings.ToUpper(word) == ac {
			return ac
		}
	}

	return strings.Title(word)
}
