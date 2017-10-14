package translator

import (
	"reflect"
	"testing"

	"github.com/jbowes/oag/openapi/v2"
)

func TestAdd(t *testing.T) {
	op := &v2.Operation{}

	out := &node{}
	out.add("/parent/child", &v2.PathItem{
		Get:     op,
		Put:     op,
		Patch:   op,
		Delete:  op,
		Head:    op,
		Options: op,
	})
	out.add("/parent/{id}", &v2.PathItem{Post: op})

	expected := &node{
		literals: []*node{{
			prefix: literal("/"),
			literals: []*node{
				{
					prefix: literal("parent"),
					literals: []*node{
						{
							prefix: literal("/"),
							literals: []*node{{
								prefix: literal("child"),
								handlers: map[string]*v2.Operation{
									"Get":     op,
									"Put":     op,
									"Patch":   op,
									"Delete":  op,
									"Head":    op,
									"Options": op,
								},
							}},
							params: []*node{{
								prefix: param("id"),
								handlers: map[string]*v2.Operation{
									"Post": op,
								},
							}},
						},
					},
				},
			},
		}},
	}

	if !reflect.DeepEqual(out, expected) {
		t.Error("got:", out, "expected:", expected)
	}
}
func TestVisit(t *testing.T) {
	trie := &node{
		literals: []*node{{
			prefix: literal("/"),
			literals: []*node{
				{
					prefix: literal("parent"),
					literals: []*node{
						{
							prefix:   literal("/"),
							literals: []*node{{prefix: literal("child")}},
							params:   []*node{{prefix: param("id")}},
						},
					},
				},
				{
					prefix: literal("sibling"),
				},
			},
		}},
	}

	var out [][]token
	for v := range trie.visit() {
		out = append(out, v.path)
	}

	nilTokenize := func(in string) []token {
		o := tokenize(in)
		return append([]token{nil}, o...)
	}

	expected := [][]token{
		nilTokenize(""),
		nilTokenize("/"),
		nilTokenize("/sibling"),
		nilTokenize("/parent"),
		nilTokenize("/parent/"),
		nilTokenize("/parent/{id}"),
		nilTokenize("/parent/child"),
	}
	if !reflect.DeepEqual(out, expected) {
		t.Error("got:", out, "expected:", expected)
	}
}

func TestTokenize(t *testing.T) {
	tcs := []struct {
		in  string
		out []token
	}{
		{"/", []token{literal("/")}},
		{"/first/thing", []token{literal("/"), literal("first"), literal("/"), literal("thing")}},
		{"/first/{param}", []token{literal("/"), literal("first"), literal("/"), param("param")}},
	}

	for _, tc := range tcs {
		t.Run(tc.in, func(t *testing.T) {
			if out := tokenize(tc.in); !tokensEqual(out, tc.out) {
				t.Error("wanted:", tc.out, "got:", out)
			}
		})
	}
}

func tokensEqual(a, b []token) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
