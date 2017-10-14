package writer

import (
	"github.com/dave/jennifer/jen"

	"github.com/jbowes/oag/pkg"
)

func defineEndpoint(f *jen.File) {
	f.Type().Id("endpoint").StructFunc(func(g *jen.Group) {
		g.Id("backend").Id("Backend")
	})
}

// defineClient defines the Client struct that contains all subclients.
func defineClient(f *jen.File, subclients []pkg.Client, prefix string) {
	f.Comment(formatComment(`
		%sClient is an API client for all endpoints.
	`, prefix))
	f.Type().Id(prefix + "Client").StructFunc(func(g *jen.Group) {
		g.Id("common").Id("endpoint").Comment(
			"Reuse a single struct instead of allocating one for each endpoint on the heap.",
		)
		g.Line()
		for _, c := range subclients {
			bare := c.Name[0 : len(c.Name)-len("Client")]
			g.Id(bare).Op("*").Id(c.Name)
		}
	})
	f.Line()

	f.Comment(formatComment(`
		New%s returns a new %sClient with the default configuration.
	`, prefix, prefix))
	f.Func().Id("New" + prefix).Params().Params(jen.Op("*").Id(prefix + "Client")).BlockFunc(func(g *jen.Group) {
		g.Id("c").Op(":=").Op("&").Id(prefix + "Client").Values()
		g.Id("c").Dot("common").Dot("backend").Op("=").Id("DefaultBackend").Call()
		g.Line()

		for _, c := range subclients {
			bare := c.Name[0 : len(c.Name)-len("Client")]
			g.Id("c").Dot(bare).Op("=").Parens(jen.Op("*").Id(c.Name)).Parens(jen.Op("&").Id("c").Dot("common"))
		}
		g.Line()

		g.Return(jen.Id("c"))
	})
}

func defineBackend(f *jen.File, prefix string) {
	newReqSig := jen.Id("NewRequest").Params(
		jen.Id("method"),
		jen.Id("path").Id("string"),
		jen.Id("query").Qual("net/url", "Values"),
		jen.Id("body").Interface(),
	).Params(jen.Op("*").Qual("net/http", "Request"), jen.Error())

	doSig := jen.Id("Do").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("request").Op("*").Qual("net/http", "Request"),
		jen.Id("v").Interface(),
		jen.Id("errFn").Func().Params(jen.Int()).Params(jen.Error()),
	).Params(
		jen.Op("*").Qual("net/http", "Response"),
		jen.Error(),
	)

	f.Comment(formatComment(`
		Backend defines the low-level interface for communicating with the remote api.
	`))
	f.Type().Id("Backend").Interface(
		newReqSig.Clone(),
		doSig.Clone(),
	)

	f.Comment(formatComment(`
		DefaultBackend returns an instance of the default Backend configuration.
	`))
	f.Func().Id("DefaultBackend").Params().Params(jen.Id("Backend")).Block(
		jen.Return(jen.Op("&").Id("defaultBackend").Values(
			jen.Id("client").Op(":").Op("&").Qual("net/http", "Client").Values(),
			jen.Id("base").Op(":").Id("base"+prefix+"URL"),
		)),
	)

	f.Type().Id("defaultBackend").Struct(
		jen.Id("client").Op("*").Qual("net/http", "Client"),
		jen.Id("base").String(),
	)

	f.Func().Params(jen.Id("b").Op("*").Id("defaultBackend")).Add(newReqSig.Clone()).BlockFunc(
		defineNewRequest,
	)
	f.Line()

	f.Func().Params(jen.Id("b").Op("*").Id("defaultBackend")).Add(doSig.Clone()).BlockFunc(
		defineDo,
	)
}

func defineNewRequest(g *jen.Group) {
	g.Var().Id("buf").Qual("bytes", "Buffer")
	g.If(jen.Id("body").Op("!=").Nil()).Block(
		jen.Id("enc").Op(":=").Qual("encoding/json", "NewEncoder").Call(jen.Op("&").Id("buf")),
		jen.If(jen.Err().Op(":=").Id("enc").Dot("Encode").Call(jen.Id("body")), jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		),
	)
	g.Line()

	g.Id("url").Op(":=").Id("b").Dot("base")
	g.If(jen.Id("path").Index(jen.Lit(0)).Op("!=").LitRune('/')).Block(
		jen.Id("url").Op("+=").Lit("/"),
	)
	g.Id("url").Op("+=").Id("path")
	g.If(jen.Id("q").Op(":=").Id("query").Dot("Encode").Call(), jen.Id("q").Op("!=").Lit("")).Block(
		jen.Id("url").Op("+=").Lit("?").Op("+").Id("q"),
	)
	g.Line()

	g.List(jen.Id("req"), jen.Err()).Op(":=").Qual("net/http", "NewRequest").Call(
		jen.Id("method"), jen.Id("url"), jen.Op("&").Id("buf"),
	)
	g.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Nil(), jen.Err()),
	)
	g.Line()

	g.If(jen.Id("body").Op("!=").Nil()).Block(
		jen.Id("req").Dot("Header").Dot("Set").Call(jen.Lit("Content-Type"), jen.Lit("application/json")),
	)
	g.Line()

	g.Return(jen.Id("req"), jen.Nil())
}

func defineDo(g *jen.Group) {
	g.Id("request").Op("=").Id("request").Dot("WithContext").Call(jen.Id("ctx"))
	g.Line()

	g.List(jen.Id("resp"), jen.Err()).Op(":=").Id("b").Dot("client").Dot("Do").Call(jen.Id("request"))
	g.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Nil(), jen.Err()),
	)
	g.Line()

	g.Defer().Id("resp").Dot("Body").Dot("Close").Call()
	g.Line()

	// XXX get status code onto error type
	g.If(jen.Id("resp").Dot("StatusCode").Op(">=").Lit(300)).BlockFunc(func(g *jen.Group) {
		// XXX return generic error type
		g.If(jen.Id("errFn").Op("==").Nil()).Block(jen.Return(jen.Nil(), jen.Nil()))
		g.Line()

		g.Id("apiErr").Op(":=").Id("errFn").Call(jen.Id("resp").Dot("StatusCode"))
		// XXX return generic error type here
		g.If(jen.Id("apiErr").Op("==").Nil()).Block(jen.Return(jen.Nil(), jen.Nil()))
		g.Line()

		g.Id("dec").Op(":=").Qual("encoding/json", "NewDecoder").Call(jen.Id("resp").Dot("Body"))
		g.If(jen.Err().Op(":=").Id("dec").Dot("Decode").Call(jen.Id("apiErr")), jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		)

		g.Return(jen.Nil(), jen.Id("apiErr"))
	})
	g.Line()

	g.If(jen.Id("v").Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
		g.Id("dec").Op(":=").Qual("encoding/json", "NewDecoder").Call(jen.Id("resp").Dot("Body"))
		g.If(jen.Err().Op(":=").Id("dec").Dot("Decode").Call(jen.Id("v")), jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		)
	})
	g.Line()

	g.Return(jen.Id("resp"), jen.Nil())
}
