package translator

import (
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
