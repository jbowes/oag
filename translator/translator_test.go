package translator

import (
	"reflect"
	"testing"

	"github.com/jbowes/oag/openapi/v2"
)

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
