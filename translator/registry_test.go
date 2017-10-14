package translator

import (
	"reflect"
	"testing"

	"github.com/jbowes/oag/openapi/v2"
	"github.com/jbowes/oag/pkg"
)

func TestConvertSchema(t *testing.T) {
	tcs := []struct {
		name string
		in   v2.Schema
		td   *pkg.TypeDecl
		out  pkg.Type
		reg  []pkg.TypeDecl
	}{
		{"string", &v2.StringSchema{}, nil, &pkg.IdentType{Name: "string"}, nil},
		{"int", &v2.IntegerSchema{}, nil, &pkg.IdentType{Name: "int"}, nil},
		{"float64", &v2.NumberSchema{}, nil, &pkg.IdentType{Name: "float64"}, nil},
		{"bool", &v2.BooleanSchema{}, nil, &pkg.IdentType{Name: "bool"}, nil},
		{"reference", &v2.ReferenceSchema{Reference: "#/definitions/foo"}, nil,
			&pkg.IdentType{Name: "Foo"}, nil},
		{
			"array",
			&v2.ArraySchema{Items: &v2.BooleanSchema{}},
			&pkg.TypeDecl{Name: "Foo"},
			&pkg.SliceType{Type: &pkg.IdentType{Name: "bool"}},
			nil,
		},
		{
			"object",
			&v2.ObjectSchema{
				Properties: &v2.SchemaMap{
					{Name: "field", Schema: &v2.StringSchema{}},
				},
				Required: &[]string{"field"},
			},
			&pkg.TypeDecl{Name: "Foo"},
			&pkg.IdentType{Name: "Foo"},
			[]pkg.TypeDecl{{Name: "Foo", Type: &pkg.StructType{
				Fields: []pkg.Field{{
					ID:   "Field",
					Type: &pkg.IdentType{Name: "string"},
					Orig: "field",
				}},
			}}},
		},
		{
			"object optional field",
			&v2.ObjectSchema{
				Properties: &v2.SchemaMap{
					{Name: "field", Schema: &v2.StringSchema{}},
				},
			},
			&pkg.TypeDecl{Name: "Foo"},
			&pkg.IdentType{Name: "Foo"},
			[]pkg.TypeDecl{{Name: "Foo", Type: &pkg.StructType{
				Fields: []pkg.Field{{
					ID:      "Field",
					Type:    &pkg.PointerType{Type: &pkg.IdentType{Name: "string"}},
					Orig:    "field",
					Comment: "Optional",
				}},
			}}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tr := &typeRegistry{}
			out := tr.convertSchema(tc.in, tc.td, false)
			if !reflect.DeepEqual(out, tc.out) {
				t.Error("got:", out, "expected:", tc.out)
			}
			if !reflect.DeepEqual(tr.types, tc.reg) {
				t.Error("got:", tr.types, "expected:", tc.reg)
			}
		})
	}
}

func TestTypeForParameter(t *testing.T) {
	tcs := []struct {
		name string
		in   v2.Parameter
		out  pkg.Type
	}{
		{"string", &v2.StringParameter{}, &pkg.IdentType{Name: "string"}},
		{"number", &v2.NumberParameter{}, &pkg.IdentType{Name: "float64"}},
		{"integer", &v2.IntegerParameter{}, &pkg.IdentType{Name: "int"}},
		{"bool", &v2.BooleanParameter{}, &pkg.IdentType{Name: "bool"}},

		{"string array",
			&v2.ArrayParameter{ArrayItem: v2.ArrayItem{Items: &v2.StringItem{}}},
			&pkg.SliceType{Type: &pkg.IdentType{Name: "string"}}},
		{"number array", &v2.ArrayParameter{ArrayItem: v2.ArrayItem{Items: &v2.NumberItem{}}},
			&pkg.SliceType{Type: &pkg.IdentType{Name: "float64"}}},
		{"integer array", &v2.ArrayParameter{ArrayItem: v2.ArrayItem{Items: &v2.IntegerItem{}}},
			&pkg.SliceType{Type: &pkg.IdentType{Name: "int"}}},
		{"bool array", &v2.ArrayParameter{ArrayItem: v2.ArrayItem{Items: &v2.BooleanItem{}}},
			&pkg.SliceType{Type: &pkg.IdentType{Name: "bool"}}},

		{
			"nested array",
			&v2.ArrayParameter{
				ArrayItem: v2.ArrayItem{Items: &v2.ArrayItem{Items: &v2.StringItem{}}},
			},
			&pkg.SliceType{Type: &pkg.SliceType{Type: &pkg.IdentType{Name: "string"}}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tr := &typeRegistry{}
			out, _ := tr.typeForParameter(tc.in)
			if !reflect.DeepEqual(out, tc.out) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}

func TestStringFormatTypeFor(t *testing.T) {
	tr := &typeRegistry{strFmt: stringFormat{"reg": "github.com/jbowes/oag.Reg"}}

	tcs := []struct {
		name string
		in   string
		out  pkg.Type
	}{
		{"no format", "", &pkg.IdentType{Name: "string"}},
		{"format not registered", "unreg", &pkg.IdentType{Name: "string"}},
		{"format registered", "reg", &pkg.IdentType{
			Name:      "Reg",
			Qualifier: "github.com/jbowes/oag",
			Marshal:   true,
		}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var f *string
			if tc.in != "" {
				f = &tc.in
			}

			out := tr.strFmt.typeFor(f)
			if !reflect.DeepEqual(out, tc.out) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}
