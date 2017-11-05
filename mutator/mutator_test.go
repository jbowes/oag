package mutator

import (
	"reflect"
	"testing"

	"github.com/jbowes/oag/pkg"
)

func TestCombineErrorsWithDefault(t *testing.T) {
	tcs := []struct {
		name string
		in   map[int]pkg.Type
		out  map[int]pkg.Type
	}{
		{"no errors", nil, nil},
		{"no default",
			map[int]pkg.Type{400: &pkg.IdentType{Name: "Error"}, 401: &pkg.IdentType{Name: "Error"}},
			map[int]pkg.Type{400: &pkg.IdentType{Name: "Error"}, 401: &pkg.IdentType{Name: "Error"}},
		},
		{"no overlap",
			map[int]pkg.Type{-1: &pkg.IdentType{Name: "Other"}, 401: &pkg.IdentType{Name: "Error"}},
			map[int]pkg.Type{-1: &pkg.IdentType{Name: "Other"}, 401: &pkg.IdentType{Name: "Error"}},
		},
		{"combined",
			map[int]pkg.Type{-1: &pkg.IdentType{Name: "Error"}, 401: &pkg.IdentType{Name: "Error"}},
			map[int]pkg.Type{-1: &pkg.IdentType{Name: "Error"}},
		},
		{"combined and not",
			map[int]pkg.Type{-1: &pkg.IdentType{Name: "Other"}, 400: &pkg.IdentType{Name: "Other"}, 401: &pkg.IdentType{Name: "Error"}},
			map[int]pkg.Type{-1: &pkg.IdentType{Name: "Other"}, 401: &pkg.IdentType{Name: "Error"}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			p := &pkg.Package{Clients: []pkg.Client{{Methods: []pkg.Method{{Errors: tc.in}}}}}
			out := combineErrorsWithDefault(p)
			outErrs := out.Clients[0].Methods[0].Errors

			if !reflect.DeepEqual(outErrs, tc.out) {
				t.Error("got:", outErrs, "expected:", tc.out)
			}
		})
	}
}

func TestInlinePrimitiveTypes(t *testing.T) {
	tcs := []struct {
		name string
		in   pkg.Package
		out  pkg.Package
	}{
		{"empty", pkg.Package{}, pkg.Package{}},
		{"inline in struct decl",
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{
					{Name: "ID", Type: &pkg.IdentType{Name: "string"}},
					{Name: "Thing", Type: &pkg.StructType{
						Fields: []pkg.Field{{Type: &pkg.IdentType{Name: "ID"}}},
					}},
				},
			},
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{
					{Name: "ID", Type: &pkg.IdentType{Name: "string"}},
					{Name: "Thing", Type: &pkg.StructType{
						Fields: []pkg.Field{{Type: &pkg.IdentType{Name: "string"}}},
					}},
				},
			},
		},

		{"inline in iter return",
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{
					{Name: "ID", Type: &pkg.IdentType{Name: "string"}},
				},
				Iters: []pkg.Iter{
					{Name: "Thing", Return: &pkg.IdentType{Name: "ID"}},
				},
			},
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{
					{Name: "ID", Type: &pkg.IdentType{Name: "string"}},
				},
				Iters: []pkg.Iter{
					{Name: "Thing", Return: &pkg.IdentType{Name: "string"}},
				},
			},
		},

		{"inline in nested decl",
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{
					{Name: "ID", Type: &pkg.IdentType{Name: "string"}},
					{Name: "OtherID", Type: &pkg.IdentType{Name: "ID"}},
					{Name: "Thing", Type: &pkg.StructType{
						Fields: []pkg.Field{{Type: &pkg.IdentType{Name: "OtherID"}}},
					}},
				},
			},
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{
					{Name: "ID", Type: &pkg.IdentType{Name: "string"}},
					{Name: "OtherID", Type: &pkg.IdentType{Name: "string"}},
					{Name: "Thing", Type: &pkg.StructType{
						Fields: []pkg.Field{{Type: &pkg.IdentType{Name: "string"}}},
					}},
				},
			},
		},

		{"inline param",
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{{Name: "ID", Type: &pkg.IdentType{Name: "string"}}},
				Clients: []pkg.Client{{Methods: []pkg.Method{{
					Params: []pkg.Param{{Type: &pkg.IdentType{Name: "ID"}}},
				}}}},
			},
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{{Name: "ID", Type: &pkg.IdentType{Name: "string"}}},
				Clients: []pkg.Client{{Methods: []pkg.Method{{
					Params: []pkg.Param{{Type: &pkg.IdentType{Name: "string"}}},
				}}}},
			},
		},

		{"inline return",
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{{Name: "ID", Type: &pkg.IdentType{Name: "string"}}},
				Clients: []pkg.Client{{Methods: []pkg.Method{{
					Return: []pkg.Type{&pkg.IdentType{Name: "ID"}},
				}}}},
			},
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{{Name: "ID", Type: &pkg.IdentType{Name: "string"}}},
				Clients: []pkg.Client{{Methods: []pkg.Method{{
					Return: []pkg.Type{&pkg.IdentType{Name: "string"}},
				}}}},
			},
		},

		{"inline error",
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{{Name: "ID", Type: &pkg.IdentType{Name: "string"}}},
				Clients: []pkg.Client{{Methods: []pkg.Method{{
					Errors: map[int]pkg.Type{-1: &pkg.IdentType{Name: "ID"}},
				}}}},
			},
			pkg.Package{
				TypeDecls: []pkg.TypeDecl{{Name: "ID", Type: &pkg.IdentType{Name: "string"}}},
				Clients: []pkg.Client{{Methods: []pkg.Method{{
					Errors: map[int]pkg.Type{-1: &pkg.IdentType{Name: "string"}},
				}}}},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := inlinePrimitiveTypes(&tc.in)
			if !reflect.DeepEqual(*out, tc.out) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}

func TestInlineResponseStruct(t *testing.T) {
	basePkg := func(c ...pkg.Client) pkg.Package {
		return pkg.Package{
			TypeDecls: []pkg.TypeDecl{
				{Name: "FieldThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "string"}}},
				}},
				{Name: "StructThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "FieldThing"}}},
				}},
			},
			Clients: c,
		}
	}

	inlinePkg := func(c ...pkg.Client) pkg.Package {
		return pkg.Package{
			TypeDecls: []pkg.TypeDecl{
				{Name: "FieldThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "string"}}},
				}},
				{Name: "StructThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.StructType{
						Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "string"}}},
					}}},
				}},
			},
			Clients: c,
		}
	}

	embeddedStructPkg := func(c ...pkg.Client) pkg.Package {
		return pkg.Package{
			TypeDecls: []pkg.TypeDecl{
				{Name: "FieldThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "string"}}},
				}},
				{Name: "StructThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{Type: &pkg.IdentType{Name: "FieldThing"}}},
				}},
			},
			Clients: c,
		}
	}

	doublePkg := func(c ...pkg.Client) pkg.Package {
		return pkg.Package{
			TypeDecls: []pkg.TypeDecl{
				{Name: "FieldThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "string"}}},
				}},
				{Name: "StructThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "FieldThing"}}},
				}},
				{Name: "OtherStructThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "FieldThing"}}},
				}},
			},
			Clients: c,
		}
	}

	inlineDoublePkg := func(c ...pkg.Client) pkg.Package {
		return pkg.Package{
			TypeDecls: []pkg.TypeDecl{
				{Name: "FieldThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "string"}}},
				}},
				{Name: "StructThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.StructType{
						Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "string"}}},
					}}},
				}},
				{Name: "OtherStructThing", Type: &pkg.StructType{
					Fields: []pkg.Field{{ID: "Field", Type: &pkg.IdentType{Name: "FieldThing"}}},
				}},
			},
			Clients: c,
		}
	}

	tcs := []struct {
		name string
		in   pkg.Package
		out  pkg.Package
	}{
		{"empty", pkg.Package{}, pkg.Package{}},

		{"not inlined if not in return or error",
			basePkg(pkg.Client{Methods: []pkg.Method{{
				Params: []pkg.Param{{Type: &pkg.IdentType{Name: "StructThing"}}},
			}}}),
			basePkg(pkg.Client{Methods: []pkg.Method{{
				Params: []pkg.Param{{Type: &pkg.IdentType{Name: "StructThing"}}},
			}}}),
		},

		{"not inlined if embedded struct",
			embeddedStructPkg(pkg.Client{Methods: []pkg.Method{{
				Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
			}}}),
			embeddedStructPkg(pkg.Client{Methods: []pkg.Method{{
				Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
			}}}),
		},

		{"inlined in return",
			basePkg(pkg.Client{Methods: []pkg.Method{{
				Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
			}}}),
			inlinePkg(pkg.Client{Methods: []pkg.Method{{
				Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
			}}}),
		},

		{"inlined in error",
			basePkg(pkg.Client{Methods: []pkg.Method{{
				Errors: map[int]pkg.Type{-1: &pkg.IdentType{Name: "StructThing"}},
			}}}),
			inlinePkg(pkg.Client{Methods: []pkg.Method{{
				Errors: map[int]pkg.Type{-1: &pkg.IdentType{Name: "StructThing"}},
			}}}),
		},

		{"inlined if used from many calls",
			basePkg(
				pkg.Client{Methods: []pkg.Method{{
					Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
				}}},
				pkg.Client{Methods: []pkg.Method{{
					Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
				}}},
			),
			inlinePkg(
				pkg.Client{Methods: []pkg.Method{{
					Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
				}}},
				pkg.Client{Methods: []pkg.Method{{
					Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
				}}},
			),
		},

		{"not inlined if used in many structs",
			doublePkg(pkg.Client{Methods: []pkg.Method{{Return: []pkg.Type{
				&pkg.IdentType{Name: "StructThing"},
				&pkg.IdentType{Name: "OtherStructThing"},
			}}}}),
			doublePkg(pkg.Client{Methods: []pkg.Method{{Return: []pkg.Type{
				&pkg.IdentType{Name: "StructThing"},
				&pkg.IdentType{Name: "OtherStructThing"},
			}}}}),
		},

		{"inlined if used in many structs but only one reachable",
			doublePkg(pkg.Client{Methods: []pkg.Method{{
				Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
			}}}),
			inlineDoublePkg(pkg.Client{Methods: []pkg.Method{{
				Return: []pkg.Type{&pkg.IdentType{Name: "StructThing"}},
			}}}),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := inlineResponseStructs(&tc.in)
			if !reflect.DeepEqual(*out, tc.out) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}
