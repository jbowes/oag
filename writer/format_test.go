package writer

import "testing"

func TestFormatComment(t *testing.T) {
	tcs := []struct {
		in, out string
	}{
		{`
			A raw comment
		`, "// A raw comment",
		},
		{`
			Multiline
			Comment
		`, "// Multiline\n// Comment",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.in, func(t *testing.T) {
			if out := formatComment(tc.in); out != tc.out {
				t.Fatal("wanted:", tc.out, "got:", out)
			}
		})
	}
}
