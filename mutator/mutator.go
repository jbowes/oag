package mutator

import (
	"github.com/jbowes/oag/pkg"
)

// Mutate runs all registered and configured mutations on the provided
// pkg.Package
func Mutate(p *pkg.Package) *pkg.Package {
	mutators := []func(*pkg.Package) *pkg.Package{
		combineErrorsWithDefault,
		inlinePrimitiveTypes,
		inlineResponseStructs,
		hoistEmbeddedStuctFields,
		removeUnusedDecls,
	}

	for _, m := range mutators {
		p = m(p)
	}
	return p
}

// combineErrorsWithDefault examines all client methods, and removes any error
// code cases that have the same type as the default, if it exists.
func combineErrorsWithDefault(p *pkg.Package) *pkg.Package {
	for _, c := range p.Clients {
		for _, m := range c.Methods {
			det, ok := m.Errors[-1]
			if !ok {
				continue
			}

			for code, et := range m.Errors {
				if code == -1 {
					continue
				}

				if det.Equal(et) {
					delete(m.Errors, code)
				}
			}
		}
	}

	return p
}

// inlinePrimitiveTypes takes any non-struct, non-interface type declarations
// and inlines their use as the original type in function parameters, return
// types, and structs.
//
// Removal of the type declaration is handled in subsequent mutations.
func inlinePrimitiveTypes(p *pkg.Package) *pkg.Package {
	for _, d := range p.TypeDecls {
		switch d.Type.(type) {
		case *pkg.StructType, *pkg.InterfaceType:
			continue
		}

		p = replaceType(p, &pkg.IdentType{Name: d.Name}, d.Type)
	}

	return p
}

func replaceType(p *pkg.Package, old, new pkg.Type) *pkg.Package {
	return walkTypes(p, func(t pkg.Type, _ typeContext) pkg.Type {
		if old.Equal(t) {
			return new
		}
		return t
	})
}

// inlineResponseStructs inlines any struct declarations for fields in structs
// used only as returns or errors, if the nested type is not used elsewhere,
// including as the field of another return or error.
func inlineResponseStructs(p *pkg.Package) *pkg.Package {
	ctxs := make(map[pkg.IdentType]struct {
		c typeContext
		n int
	}, len(p.TypeDecls))

	reachDecls(p, func(c typeContext, i *pkg.IdentType) bool {
		v := ctxs[*i]
		v.c |= c
		v.n++
		ctxs[*i] = v
		return true
	})

	for i, c := range ctxs {
		if c.c&methodParam > 0 {
			delete(ctxs, i)
			continue
		}
	}

	for _, d := range p.TypeDecls {
		di := pkg.IdentType{Name: d.Name}
		if pc, ok := ctxs[di]; ok {
			d.Type = recurseType(d.Type, decl, func(t pkg.Type, c typeContext) pkg.Type {
				if t == d.Type {
					return t
				}

				if c&embeddedStruct > 0 {
					return t
				}

				if i, ok := t.(*pkg.IdentType); ok {
					if cc, ok := ctxs[*i]; ok && pc.n >= cc.n {
						return resolve(p, i)
					}
				}

				return t
			})
		}
	}

	return p
}

// hoistEmbeddedStuctFields finds any embedded structs from allOf schemas, and
// hoists their fields into the embedding struct, when the embedded struct is not
// referenced elsewhere.
func hoistEmbeddedStuctFields(p *pkg.Package) *pkg.Package {
	// XXX dedupe this code with inlineResponseStructs
	ctxs := make(map[pkg.IdentType]struct {
		c typeContext
		n int
	}, len(p.TypeDecls))

	reachDecls(p, func(c typeContext, i *pkg.IdentType) bool {
		v := ctxs[*i]
		v.c |= c
		v.n++
		ctxs[*i] = v
		return true
	})

	for _, d := range p.TypeDecls {
		di := pkg.IdentType{Name: d.Name}
		if pc, ok := ctxs[di]; ok {
			d.Type = recurseType(d.Type, decl, func(t pkg.Type, c typeContext) pkg.Type {
				st, ok := t.(*pkg.StructType)
				if !ok {
					return t
				}

				var nf []pkg.Field

				for i := range st.Fields {
					f := st.Fields[i]
					if f.ID != "" {
						nf = append(nf, f)
						continue
					}

					var embedded *pkg.StructType
					switch ft := f.Type.(type) {
					case *pkg.IdentType:
						if cc, ok := ctxs[*ft]; !ok || pc.n < cc.n {
							nf = append(nf, f)
							continue
						}

						resolved := resolve(p, ft)
						if rst, ok := resolved.(*pkg.StructType); ok {
							embedded = rst
						} else {
							nf = append(nf, f)
							continue
						}
					case *pkg.StructType:
						embedded = ft
					default:
						nf = append(nf, f)
						continue
					}

					for j := range embedded.Fields {
						nf = append(nf, embedded.Fields[j])
					}
				}

				st.Fields = nf
				return st
			})
		}
	}

	return p
}

type typeSet map[pkg.IdentType]struct{}

func (t typeSet) Mark(i pkg.IdentType) {
	t[i] = struct{}{}
}

func (t typeSet) Seen(i pkg.IdentType) bool {
	_, ok := t[i]
	return ok
}

// removeUnusedDecls removes any type declarations that aren't referenced elsewhere
// in the generated code.
func removeUnusedDecls(p *pkg.Package) *pkg.Package {
	ts := make(typeSet, len(p.TypeDecls))

	reachDecls(p, func(_ typeContext, i *pkg.IdentType) bool {
		ret := !ts.Seen(*i)
		ts.Mark(*i)
		return ret
	})

	swept := make([]pkg.TypeDecl, 0, len(p.TypeDecls))
	for _, d := range p.TypeDecls {
		if ts.Seen(pkg.IdentType{Name: d.Name}) {
			swept = append(swept, d)
		}
	}
	p.TypeDecls = swept

	sweptIter := make([]pkg.Iter, 0, len(p.Iters))
	for _, i := range p.Iters {
		if ts.Seen(pkg.IdentType{Name: i.Name}) {
			sweptIter = append(sweptIter, i)
		}
	}
	p.Iters = sweptIter

	return p
}

type typeContext uint16

const (
	decl typeContext = 1 << iota
	iter
	methodParam
	methodReturn
	methodError

	structField
	embeddedStruct

	// none typeContext = 0
	any typeContext = 0xFFFF
)

type stackItem struct {
	t pkg.Type
	c typeContext
}

func reachDecls(p *pkg.Package, fn func(typeContext, *pkg.IdentType) bool) {
	var stack []stackItem

	for _, c := range p.Clients {
		for _, m := range c.Methods {
			for _, param := range m.Params {
				stack = append(stack, stackItem{param.Type, methodParam})
			}

			for _, ret := range m.Return {
				stack = append(stack, stackItem{ret, methodReturn})
			}

			for _, e := range m.Errors {
				stack = append(stack, stackItem{e, methodError})
			}
		}
	}

	for len(stack) > 0 {
		item := stack[0]
		stack = stack[1:]

		eachIdent(item.t, func(i *pkg.IdentType) {
			if !fn(item.c, i) {
				return
			}

			for _, d := range p.TypeDecls {
				di := pkg.IdentType{Name: d.Name}
				if *i == di {
					eachIdent(d.Type, func(ci *pkg.IdentType) {
						stack = append(stack, stackItem{ci, item.c | decl})
					})
					return
				}
			}

			for _, pi := range p.Iters {
				ii := pkg.IdentType{Name: pi.Name}
				if *i == ii {
					eachIdent(pi.Return, func(ci *pkg.IdentType) {
						stack = append(stack, stackItem{ci, item.c | iter})
					})
					return
				}
			}
		})
	}
}

func walkTypes(p *pkg.Package, fn func(pkg.Type, typeContext) pkg.Type) *pkg.Package {
	for i, d := range p.TypeDecls {
		d.Type = recurseType(d.Type, decl, fn)
		p.TypeDecls[i] = d
	}

	for i, itr := range p.Iters {
		itr.Return = recurseType(itr.Return, iter, fn)
		p.Iters[i] = itr
	}

	for _, c := range p.Clients {
		for _, m := range c.Methods {
			for i, param := range m.Params {
				param.Type = recurseType(param.Type, methodParam, fn)
				m.Params[i] = param
			}

			for i, ret := range m.Return {
				m.Return[i] = recurseType(ret, methodReturn, fn)
			}

			for k, e := range m.Errors {
				m.Errors[k] = recurseType(e, methodError, fn)
			}
		}
	}

	return p
}

func recurseType(typ pkg.Type, parentCtx typeContext, fn func(pkg.Type, typeContext) pkg.Type) pkg.Type {
	switch t := typ.(type) {
	case *pkg.StructType:
		for i := range t.Fields {
			c := structField
			if t.Fields[i].ID == "" {
				c = embeddedStruct
			}

			t.Fields[i].Type = recurseType(t.Fields[i].Type, c, fn)
		}
	case *pkg.SliceType:
		t.Type = recurseType(t.Type, parentCtx, fn)
	case *pkg.IterType:
		t.Type = recurseType(t.Type, parentCtx|iter, fn)
	case *pkg.PointerType:
		t.Type = recurseType(t.Type, parentCtx, fn)
	case *pkg.InterfaceType:
		for i := range t.Methods {
			t.Methods[i].Return = recurseType(t.Methods[i].Return, parentCtx, fn)
		}

		for i := range t.Implementors {
			t.Implementors[i].Type = recurseType(t.Implementors[i].Type, parentCtx, fn)
		}
	}

	return fn(typ, parentCtx)
}

func eachIdent(typ pkg.Type, fn func(*pkg.IdentType)) {
	recurseType(typ, any, func(t pkg.Type, _ typeContext) pkg.Type {
		if i, ok := t.(*pkg.IdentType); ok {
			fn(i)
		}
		return t
	})
}

func resolve(p *pkg.Package, i *pkg.IdentType) pkg.Type {
	for _, d := range p.TypeDecls {
		di := pkg.IdentType{Name: d.Name}
		if *i == di {
			return d.Type
		}
	}

	return i
}
