package writer

import (
	"fmt"
	"go/format"
	"testing"

	"github.com/dave/jennifer/jen"

	"github.com/jbowes/oag/pkg"
)

func TestWriteType(t *testing.T) {
	tcs := []struct {
		name string
		in   pkg.Type
		out  string
	}{
		{"ident", &pkg.IdentType{Name: "string"}, "string"},
		{"qualified ident", &pkg.IdentType{Qualifier: "github.com/jbowes/whatever", Name: "String"}, "whatever.String"},
		{"slice", &pkg.SliceType{Type: &pkg.IdentType{Name: "string"}}, "[]string"},
		{"pointer", &pkg.PointerType{Type: &pkg.IdentType{Name: "string"}}, "*string"},
		{"empty struct", &pkg.StructType{}, "struct{}"},
		{"empty interface", &pkg.EmptyInterfaceType{}, "interface{}"},
		{"struct",
			&pkg.StructType{Fields: []pkg.Field{
				{ID: "Foo", Type: &pkg.IdentType{Name: "string"}},
			}},
			`struct {
				Foo string
			}`,
		},

		{"nested struct",
			&pkg.StructType{Fields: []pkg.Field{
				{ID: "Foo", Type: &pkg.IdentType{Name: "string"}},
				{ID: "Bar", Type: &pkg.StructType{}},
			}},
			`struct {
				Foo string

				Bar struct{}
			}`,
		},

		{"nested struct (first)",
			&pkg.StructType{Fields: []pkg.Field{
				{ID: "Bar", Type: &pkg.StructType{}},
			}},
			`struct {
				Bar struct{}
			}`,
		},
		{"map",
			&pkg.MapType{
				Key:   &pkg.IdentType{Name: "string"},
				Value: &pkg.IdentType{Name: "string"},
			},
			`map[string]string`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			s := jen.Type().Id("t").Do(writeType(tc.in))
			out := fmt.Sprintf("%#v", s)

			formatted, _ := format.Source([]byte("type t " + tc.out))
			wanted := string(formatted)
			if out != wanted {
				t.Error("wanted:\n", wanted, "\ngot:\n", out)
			}
		})
	}
}
