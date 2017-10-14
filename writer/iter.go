package writer

import (
	"github.com/dave/jennifer/jen"
	"github.com/gedex/inflector"

	"github.com/jbowes/oag/pkg"
)

func defineIter(f *jen.File, iter *pkg.Iter) {
	page := jen.Id("page")
	switch t := iter.Return.(type) {
	case *pkg.PointerType:
		page.Do(writeType(&pkg.SliceType{Type: t.Type}))
	default:
		page.Do(writeType(&pkg.SliceType{Type: iter.Return}))
	}
	// XXX add comment about following pagination when supported, and where appropriate.
	f.Comment(formatComment(`
		%s Iterates over a result set of %s.
	`, iter.Name, inflector.Pluralize(typeName(iter.Return))))
	f.Type().Id(iter.Name).Struct(
		page,
		jen.Id("i").Int(),
		jen.Empty(),
		jen.Err().Error(),
		jen.Id("first").Bool(),
	)

	// XXX should return something? err?
	f.Comment(formatComment(`
		Close closes the %s and releases any associated resources.
		After Close, any calls to Current will return an error.
	`, iter.Name))
	f.Func().Params(jen.Id("i").Op("*").Id(iter.Name)).Id("Close").Params().BlockFunc(func(g *jen.Group) {
	})

	f.Comment(formatComment(`
		Next advances the %s and returns a boolean indicating if the end has been reached.
		Next must be called before the first call to Current.
		Calls to Current after Next returns false will return an error.
	`, iter.Name))
	f.Func().Params(jen.Id("i").Op("*").Id(iter.Name)).Id("Next").Params().Params(jen.Bool()).BlockFunc(func(g *jen.Group) {
		g.If(jen.Id("i").Dot("first").Op("&&").Id("i").Dot("err").Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
			g.Id("i").Dot("first").Op("=").False()
			g.Return(jen.True())
		})
		g.Id("i").Dot("first").Op("=").False()
		g.Id("i").Dot("i").Op("++")
		g.Return(jen.Id("i").Dot("i").Op("<").Len(jen.Id("i").Dot("page")))
	})

	f.Comment(formatComment(`
		Current returns the current %s, and an optional error. Once an error has been returned,
		the %s is closed, or the end of iteration is reached, subsequent calls to Current
		will return an error.
	`, typeName(iter.Return), iter.Name))
	f.Func().Params(jen.Id("i").Op("*").Id(iter.Name)).Id("Current").Params().Params(jen.Do(writeType(iter.Return)), jen.Error()).BlockFunc(func(g *jen.Group) {
		// XXX handle when i is past end of slice
		g.If(jen.Id("i").Dot("err").Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
			g.Return(jen.Nil(), jen.Id("i").Dot("err"))
		})
		g.Return(jen.Op("&").Id("i").Dot("page").Index(jen.Id("i").Dot("i")), jen.Nil())
	})
}
