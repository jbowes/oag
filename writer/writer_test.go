package writer

import (
	"fmt"
	"go/format"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/jbowes/oag/pkg"
)

func TestSetQueryArgs(t *testing.T) {
	tcs := []struct {
		name string
		in   []pkg.Param
		out  string
	}{
		{"no args", nil, ""},
		{"string arg",
			[]pkg.Param{{ID: "arg"}},
			`

				q := make(url.Values)
				q.Set("arg", arg)
			`,
		},
		{"different name",
			[]pkg.Param{{ID: "arg", Orig: "arg_thing"}},
			`

				q := make(url.Values)
				q.Set("arg_thing", arg)
			`,
		},
		{"marshal",
			[]pkg.Param{{ID: "arg", Type: &pkg.IdentType{Marshal: true}}},
			`

				q := make(url.Values)
				argBytes, err := arg.MarshalText()
    			if err != nil {
    				return
    			}
    			q.Set("arg", string(argBytes))
			`,
		},
		{"space after martial, not between regular",
			[]pkg.Param{
				{ID: "arg1", Type: &pkg.IdentType{Marshal: true}},
				{ID: "arg2"},
				{ID: "arg3"},
			},
			`

				q := make(url.Values)
				arg1Bytes, err := arg1.MarshalText()
    			if err != nil {
    				return
    			}
				q.Set("arg1", string(arg1Bytes))

				q.Set("arg2", arg2)
				q.Set("arg3", arg3)
			`,
		},

		{"multi collection string arg",
			[]pkg.Param{{ID: "arg", Collection: pkg.Multi, Type: &pkg.SliceType{}}},
			`

				q := make(url.Values)
				for _, v := range arg {
					q.Add("arg", v)
				}
			`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			sqa := jen.Id("v").Op("=").Func().Params().BlockFunc(func(g *jen.Group) {
				setQueryArgs(g, nil, tc.in)
			})

			out := fmt.Sprintf("%#v", sqa)
			formatted, _ := format.Source([]byte("v = func() {" + tc.out + "}"))
			if out != string(formatted) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}

func TestSetOptQueryArgs(t *testing.T) {
	tcs := []struct {
		name string
		in   []pkg.Field
		out  string
	}{
		{"no args", nil, ""},
		{"string arg",
			[]pkg.Field{{ID: "arg", Type: &pkg.PointerType{}}},
			`

				var q url.Values
				if opts != nil {
					q = make(url.Values)
					if opts.arg != nil {
						q.Set("arg", *opts.arg)
					}
				}
			`,
		},
		{"different name",
			[]pkg.Field{{ID: "arg", Orig: "arg_thing", Type: &pkg.PointerType{}}},
			`

				var q url.Values
				if opts != nil {
					q = make(url.Values)
					if opts.arg != nil {
						q.Set("arg_thing", *opts.arg)
					}
				}
			`,
		},
		{"marshal",
			[]pkg.Field{{ID: "arg", Type: &pkg.PointerType{Type: &pkg.IdentType{Marshal: true}}}},
			`

				var q url.Values
				if opts != nil {
					q = make(url.Values)
					if opts.arg != nil {
						b, err := opts.arg.MarshalText()
						if err != nil {
							return
						}
						q.Set("arg", string(b))
					}
				}
			`,
		},
		{"space after martial, not between regular",
			[]pkg.Field{
				{ID: "arg1", Type: &pkg.PointerType{Type: &pkg.IdentType{Marshal: true}}},
				{ID: "arg2", Type: &pkg.PointerType{}},
				{ID: "arg3", Type: &pkg.PointerType{}},
			},
			`

				var q url.Values
				if opts != nil {
					q = make(url.Values)
					if opts.arg1 != nil {
						b, err := opts.arg1.MarshalText()
						if err != nil {
							return
						}
						q.Set("arg1", string(b))
					}

					if opts.arg2 != nil {
						q.Set("arg2", *opts.arg2)
					}

					if opts.arg3 != nil {
						q.Set("arg3", *opts.arg3)
					}
				}
			`,
		},

		{"multi collection string arg",
			[]pkg.Field{{ID: "arg", Collection: pkg.Multi, Type: &pkg.PointerType{Type: &pkg.SliceType{}}}},
			`

				var q url.Values
				if opts != nil {
					q = make(url.Values)
					for _, v := range opts.arg {
						q.Add("arg", v)
					}
				}
			`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			soqa := jen.Id("v").Op("=").Func().Params().BlockFunc(func(g *jen.Group) {
				setOptQueryArgs(g, nil, false, tc.in)
			})

			out := fmt.Sprintf("%#v", soqa)
			formatted, _ := format.Source([]byte("v = func() {" + tc.out + "}"))
			if out != string(formatted) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}
func TestSetHeaderArgs(t *testing.T) {
	tcs := []struct {
		name string
		in   []pkg.Param
		out  string
	}{
		{"no args", nil, ""},
		{"string arg",
			[]pkg.Param{{ID: "arg"}},
			`req.Header.Set("arg", arg)

			`,
		},
		{"different name",
			[]pkg.Param{{ID: "arg", Orig: "arg_thing"}},
			`req.Header.Set("arg_thing", arg)

			`,
		},
		{"marshal",
			[]pkg.Param{{ID: "arg", Type: &pkg.IdentType{Marshal: true}}},
			`argBytes, err := arg.MarshalText()
    			if err != nil {
    				return
    			}
    			req.Header.Set("arg", string(argBytes))

			`,
		},
		{"space after martial, not between regular",
			[]pkg.Param{
				{ID: "arg1", Type: &pkg.IdentType{Marshal: true}},
				{ID: "arg2"},
				{ID: "arg3"},
			},
			`arg1Bytes, err := arg1.MarshalText()
    			if err != nil {
    				return
    			}
				req.Header.Set("arg1", string(arg1Bytes))

				req.Header.Set("arg2", arg2)
				req.Header.Set("arg3", arg3)

			`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			sha := jen.Id("v").Op("=").Func().Params().BlockFunc(func(g *jen.Group) {
				setHeaderArgs(g, nil, tc.in)
			})

			out := fmt.Sprintf("%#v", sha)
			formatted, _ := format.Source([]byte("v = func() {" + tc.out + "}"))
			if out != string(formatted) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}

func TestErrSelectFunc(t *testing.T) {
	tcs := []struct {
		name string
		in   pkg.Method
		out  string
	}{
		{"no error codes", pkg.Method{}, "nil"},
		{"only default",
			pkg.Method{Errors: map[int]pkg.Type{
				-1: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error"}},
			}},
			`func(code int) error {
				return &Error{}
			}`,
		},
		{"many codes",
			pkg.Method{Errors: map[int]pkg.Type{
				400: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error400"}},
				401: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error401"}},
			}},
			`func(code int) error {
				switch code {
				case 400:
					return &Error400{}
				case 401:
					return &Error401{}
				default:
					return nil
				}
			}`,
		},
		{"codes and default",
			pkg.Method{Errors: map[int]pkg.Type{
				-1:  &pkg.PointerType{Type: &pkg.IdentType{Name: "Error"}},
				400: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error400"}},
				401: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error401"}},
			}},
			`func(code int) error {
				switch code {
				case 400:
					return &Error400{}
				case 401:
					return &Error401{}
				default:
					return &Error{}
				}
			}`,
		},
		{"single code no default",
			pkg.Method{Errors: map[int]pkg.Type{
				400: &pkg.PointerType{Type: &pkg.IdentType{Name: "Error400"}},
			}},
			`func(code int) error {
				if code == 400 {
					return &Error400{}
				}
				return nil
			}`,
		},
		{"same successive error groups case",
			pkg.Method{Errors: map[int]pkg.Type{
				400: &pkg.PointerType{Type: &pkg.IdentType{Name: "ErrorA"}},
				401: &pkg.PointerType{Type: &pkg.IdentType{Name: "ErrorA"}},
				403: &pkg.PointerType{Type: &pkg.IdentType{Name: "ErrorB"}},
				404: &pkg.PointerType{Type: &pkg.IdentType{Name: "ErrorA"}},
				405: &pkg.PointerType{Type: &pkg.IdentType{Name: "ErrorA"}},
			}},
			`func(code int) error {
				switch code {
				case 400, 401:
					return &ErrorA{}
				case 403:
					return &ErrorB{}
				case 404, 405:
					return &ErrorA{}
				default:
					return nil
				}
			}`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			esf := jen.Id("v").Op("=").Add(errSelectFunc(&tc.in))
			out := fmt.Sprintf("%#v", esf)
			formatted, _ := format.Source([]byte("v = " + tc.out))
			if out != string(formatted) {
				t.Error("got:", out, "expected:", tc.out)
			}
		})
	}
}
