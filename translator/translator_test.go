package translator

import (
	"reflect"
	"testing"

	"github.com/jbowes/oag/openapi/v2"
	"github.com/jbowes/oag/pkg"
)

func TestConvertOperationResponses(t *testing.T) {
	tcs := []struct {
		name string
		resp v2.Responses
		ret  []pkg.Type
		errs map[int]pkg.Type
	}{
		{
			name: "204 response only",
			resp: v2.Responses{
				Codes: map[int]v2.Response{
					204: {Schema: &v2.ObjectSchema{}},
				},
			},
			ret:  []pkg.Type{&pkg.IdentType{Name: "error"}},
			errs: make(map[int]pkg.Type),
		},
		{
			name: "200 response only",
			resp: v2.Responses{
				Codes: map[int]v2.Response{
					200: {Schema: &v2.ObjectSchema{}},
				},
			},
			ret: []pkg.Type{
				&pkg.PointerType{Type: &pkg.StructType{}},
				&pkg.IdentType{Name: "error"},
			},
			errs: make(map[int]pkg.Type),
		},
		{
			name: "2XX reference iterator response",
			resp: v2.Responses{
				Codes: map[int]v2.Response{
					200: {Schema: &v2.ArraySchema{Items: &v2.ReferenceSchema{Reference: "Fake"}}},
				},
			},
			ret: []pkg.Type{
				&pkg.IterType{Type: &pkg.PointerType{Type: &pkg.IdentType{Name: "FakeIter"}}},
			},
			errs: make(map[int]pkg.Type),
		},
		{
			name: "4XX reference response only",
			resp: v2.Responses{
				Codes: map[int]v2.Response{
					400: {Schema: &v2.ReferenceSchema{Reference: "BadRequest"}},
				},
			},
			ret: []pkg.Type{
				&pkg.IdentType{Name: "error"},
			},
			errs: map[int]pkg.Type{
				400: &pkg.PointerType{Type: &pkg.IdentType{Name: "BadRequest"}},
			},
		},
		{
			name: "Default response",
			resp: v2.Responses{
				Default: &v2.Response{
					Schema: &v2.ReferenceSchema{Reference: "Error"},
				},
			},
			ret: []pkg.Type{
				&pkg.IdentType{Name: "error"},
			},
			errs: map[int]pkg.Type{
				-1: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error"}},
			},
		},
		{
			name: "Default response",
			resp: v2.Responses{
				Default: &v2.Response{
					Reference: "Error",
				},
			},
			ret: []pkg.Type{
				&pkg.IdentType{Name: "error"},
			},
			errs: map[int]pkg.Type{
				-1: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error"}},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tr := &typeRegistry{}
			ret, errs := convertOperationResponses(nil, tr, "Get", &tc.resp, &pkg.Package{})

			if !reflect.DeepEqual(ret, tc.ret) {
				t.Error("got:", ret, "expected:", tc.ret)
			}
			if !reflect.DeepEqual(errs, tc.errs) {
				t.Error("got:", errs, "expected:", tc.errs)
			}
		})
	}
}

func TestMethodMap(t *testing.T) {
	tcs := []struct {
		name string
		in   map[string]*v2.Operation
		out  map[string]string
	}{
		{"get", map[string]*v2.Operation{"Get": nil}, map[string]string{"Get": "Get"}},
		{"post", map[string]*v2.Operation{"Post": nil}, map[string]string{"Post": "Create"}},
		{"patch", map[string]*v2.Operation{"Patch": nil}, map[string]string{"Patch": "Update"}},
		{
			"post and put",
			map[string]*v2.Operation{"Post": nil, "Put": nil},
			map[string]string{"Post": "Create", "Put": "Update"},
		},
		{
			"put and patch",
			map[string]*v2.Operation{"Put": nil, "Patch": nil},
			map[string]string{"Put": "Create", "Patch": "Update"},
		},
		{
			"post, put and patch",
			map[string]*v2.Operation{"Post": nil, "Put": nil, "Patch": nil},
			map[string]string{"Post": "Create", "Put": "Replace", "Patch": "Update"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := methodMap(tc.in)
			if !reflect.DeepEqual(out, tc.out) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}

func TestFormatReserved(t *testing.T) {
	tcs := []struct {
		name    string
		in      string
		context string
		out     string
	}{
		{"not reserved", "value", "testing", "value"},
		{"reserved", "type", "testing", "testingType"},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if out := formatReserved(tc.in, tc.context); out != tc.out {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}
