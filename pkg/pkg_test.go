package pkg

import "testing"

func TestTypeEqual(t *testing.T) {
	cases := []Type{
		&IdentType{Name: "string"},
		&IdentType{Name: "Thing", Qualifier: "github.com/jbowes/oag"},
		&PointerType{Type: &IdentType{Name: "string"}},
		&SliceType{Type: &IdentType{Name: "string"}},
		&IterType{Type: &IdentType{Name: "string"}},
		&StructType{Fields: []Field{}},
		&StructType{Fields: []Field{{ID: "Name", Type: &IdentType{Name: "string"}}}},
		&StructType{Fields: []Field{{ID: "Name", Type: &IdentType{Name: "Thing"}}}},
		&MapType{Key: &IdentType{Name: "string"}, Value: &IdentType{Name: "int"}},
		&MapType{Key: &IdentType{Name: "int"}, Value: &SliceType{&IdentType{Name: "int"}}},
		&EmptyInterfaceType{},
		&InterfaceType{Methods: []InterfaceMethod{}},
		&InterfaceType{Methods: []InterfaceMethod{
			{Name: "GetFoo", Return: &IdentType{Name: "string"}},
		}},
		&InterfaceType{Methods: []InterfaceMethod{
			{Name: "GetFoo", Return: &IdentType{Name: "int"}},
		}},
		&InterfaceType{Methods: []InterfaceMethod{
			{Name: "GetBar", Return: &IdentType{Name: "int"}},
			{Name: "GetFoo", Return: &IdentType{Name: "string"}},
		}},
	}

	for i := range cases {
		for j := range cases {
			if (i == j) != cases[i].Equal(cases[j]) {
				t.Error("bad result for", i, j)
			}
		}
	}
}
