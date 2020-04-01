package writer

import (
	"fmt"
	"strings"

	"github.com/lithammer/dedent"
)

// Jen uses multiline /* */ comments. // looks nicer. Preformat with this func.
func formatComment(comment string, args ...interface{}) string {
	comment = fmt.Sprintf(comment, args...)
	comment = dedent.Dedent(comment)
	comment = "// " + strings.TrimSpace(comment)
	return strings.Replace(comment, "\n", "\n// ", -1)
}
