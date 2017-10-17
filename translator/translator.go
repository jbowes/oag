// Package translator translates an openapi document into an API code definition
package translator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jbowes/oag/openapi/v2"
	"github.com/jbowes/oag/pkg"
)

// Translate translates an openapi.Document into a series of Packages
func Translate(doc *v2.Document, qual, name string, types map[string]string, stringFormats map[string]string) (*pkg.Package, error) {
	p := &pkg.Package{
		Qualifier: qual,
		Name:      name,
		BaseURL:   "https://" + *doc.Host + *doc.BasePath,
	}

	tr := &typeRegistry{strFmt: stringFormats}
	if doc.Definitions != nil {
		for _, def := range *doc.Definitions {
			convertDefinition(tr, def.Name, def.Schema, types)
		}
	}

	trie := &node{}
	for path, pi := range doc.Paths {
		p := pi
		trie.add(path, &p)
	}

	clients := make(map[string]*pkg.Client)
	getClient := func(path []token) *pkg.Client {
		clientName := formatID(path[2].value(), "Client")
		if len(path) == 3 { // Define the client
			clients[clientName] = &pkg.Client{
				Name:    clientName,
				Comment: fmt.Sprintf("%s provides access to the /%s APIs", clientName, path[2].value()),
			}
		}

		return clients[clientName]
	}

	for n := range trie.visit() {
		if len(n.path) < 3 {
			continue
		}

		client := getClient(n.path)

		mm := methodMap(n.n.handlers)
		for m, o := range n.n.handlers {
			method := convertOperation(tr, doc, n, m, mm[m], o, client, p)
			client.Methods = append(client.Methods, *method)

		}
		sort.Slice(client.Methods, func(i, j int) bool { return client.Methods[i].Name < client.Methods[j].Name })
	}

	for _, c := range clients {
		p.Clients = append(p.Clients, *c)
	}
	sort.Slice(p.Clients, func(i, j int) bool { return p.Clients[i].Name < p.Clients[j].Name })

	sort.Slice(p.Iters, func(i, j int) bool { return p.Iters[i].Name < p.Iters[j].Name })

	p.TypeDecls = tr.types
	sort.Slice(p.TypeDecls, func(i, j int) bool { return p.TypeDecls[i].Name < p.TypeDecls[j].Name })

	return p, nil
}

func convertDefinition(tr *typeRegistry, name string, def v2.Schema, types map[string]string) {
	dataName := formatID(name)
	comment := fmt.Sprintf("%s is a data type for API communication.", dataName)
	if title := def.GetTitle(); title != nil {
		comment += "\n\n" + *title
	}
	if description := def.GetDescription(); description != nil {
		comment += "\n" + *description
	}

	if mapping, ok := types[name]; ok {
		idx := strings.LastIndex(mapping, ".")
		tr.add(pkg.TypeDecl{
			Name:    dataName,
			Comment: comment,
			Type: &pkg.IdentType{
				Qualifier: mapping[:idx],
				Name:      mapping[idx+1:],
			},
		})

		return
	}

	tr.convertSchema(def, &pkg.TypeDecl{
		Name:    dataName,
		Comment: comment,
	}, true)
}

func convertOperation(tr *typeRegistry, def *v2.Document, n *visited, httpMethod, prefix string, o *v2.Operation, client *pkg.Client, p *pkg.Package) *pkg.Method {
	// if array response, change Get to List
	if httpMethod == "Get" {
		for code, r := range o.Responses.Codes {
			if code >= 200 && code < 300 {
				if _, ok := r.Schema.(*v2.ArraySchema); ok {
					prefix = "List"
					break
				}
			}
		}
	}

	var params []param
	parts := []string{prefix}
	for _, p := range n.path[3:] {
		if t, ok := p.(param); ok {
			params = append(params, t)
			continue
		}
		if p.value() == "/" {
			continue
		}
		parts = append(parts, p.value())
	}
	methodName := formatID(parts...)

	var path string
	var docPath string
	for _, p := range n.path[1:] {
		if pp, ok := p.(param); ok {
			path += "%s" // XXX check type
			docPath += ":" + string(pp)
		} else {
			path += p.value()
			docPath += p.value()
		}
	}

	comment := fmt.Sprintf("%s corresponds to the %s %s endpoint.", methodName, strings.ToUpper(httpMethod), docPath)
	if o.Summary != nil {
		comment += "\n\n" + *o.Summary
	}
	if o.Description != nil {
		comment += "\n" + *o.Description
	}

	method := &pkg.Method{
		Comment:    comment,
		Name:       methodName,
		Path:       path,
		HTTPMethod: httpMethod,
	}
	method.Receiver.ID = "c"
	method.Receiver.Type = client.Name

	// add arguments. logic should be:
	// path parameters in order of path
	// required non body arguments
	// then body
	// then optional args in a struct
	var pathParams []pkg.Param // always first
	var body *pkg.Param        // always last required arg, before opts
	var opts []pkg.Field       // always last

	for _, p := range params {
		pathParams = append(pathParams, pkg.Param{
			ID:   string(p),
			Kind: pkg.Path,
		})
	}

	for _, p := range o.Parameters {
		newBody, newParam, newOpts := convertParameter(tr, def, methodName, pathParams, p)
		if newBody != nil {
			body = newBody
		}
		method.Params = append(method.Params, newParam...)
		opts = append(opts, newOpts...)
	}

	method.Params = append(pathParams, method.Params...)
	if body != nil {
		method.Params = append(method.Params, *body)
	}

	if len(opts) > 0 {
		optsName := methodName + "Opts"
		if len(parts) == 1 {
			optsName = formatID(n.path[2].value(), methodName, "Opts")
		}
		// XXX dedupe similar Opts structs
		td := pkg.TypeDecl{
			Name:    optsName,
			Comment: fmt.Sprintf("%s holds optional argument values", optsName),
			Type: &pkg.StructType{
				Fields: opts,
			},
		}
		tr.add(td)

		method.Params = append(method.Params, pkg.Param{
			ID:   "opts",
			Type: &pkg.PointerType{Type: &pkg.IdentType{Name: optsName}},
			Kind: pkg.Opts,
		})
	}

	method.Return, method.Errors = convertOperationResponses(def, tr, methodName, o, p)

	return method
}

func convertOperationResponses(doc *v2.Document, tr *typeRegistry, methodName string, o *v2.Operation, p *pkg.Package) ([]pkg.Type, map[int]pkg.Type) {
	var rets []pkg.Type
	errs := make(map[int]pkg.Type)

	iter := false
	for code, r := range o.Responses.Codes {
		switch {
		case code < 200: // XXX should these be handled?
		case code == 204: // no response
		case code < 300: // XXX handle multiple 2XX returns
			ret := tr.convertSchema(r.Schema, &pkg.TypeDecl{
				Name: methodName + "Response",
			}, false)

			if it, ok := ret.(*pkg.SliceType); ok {
				iter = true
				name := it.Type.(*pkg.IdentType).Name + "Iter"
				ret = &pkg.IterType{Type: &pkg.PointerType{
					Type: &pkg.IdentType{Name: name},
				}}

				p.Iters = append(p.Iters, pkg.Iter{
					Name:   name,
					Return: &pkg.PointerType{Type: it.Type},
				})

			} else {
				ret = &pkg.PointerType{Type: ret}
			}

			rets = append(rets, ret)
		default:
			if r.Reference != "" {
				parts := strings.Split(r.Reference, "/")
				refname := parts[len(parts)-1]
				r = (*doc.Responses)[refname]
			}

			if t, ok := r.Schema.(*v2.ReferenceSchema); ok {
				parts := strings.Split(t.Reference, "/")
				refname := parts[len(parts)-1] // XXX rename for snake case

				errs[code] = &pkg.PointerType{
					Type: &pkg.IdentType{Name: refname},
				}
			}
		}
	}

	// XXX handle when this is the success case (no 2XX responses)
	if o.Responses.Default != nil {
		if t, ok := o.Responses.Default.Schema.(*v2.ReferenceSchema); ok {
			parts := strings.Split(t.Reference, "/")
			refname := parts[len(parts)-1] // XXX rename for snake case

			errs[-1] = &pkg.PointerType{
				Type: &pkg.IdentType{Name: refname},
			}
		}

	}

	if !iter {
		rets = append(rets, &pkg.IdentType{Name: "error"})
	}

	return rets, errs
}

func convertParameter(tr *typeRegistry, def *v2.Document, methodName string, pathParams []pkg.Param, p v2.Parameter) (*pkg.Param, []pkg.Param, []pkg.Field) {
	if ref, ok := p.(*v2.ReferenceParamter); ok {
		parts := strings.Split(ref.Reference, "/")
		p = (*def.Parameters)[parts[len(parts)-1]]
	}

	switch p.GetIn() {
	case "body":
		// body will always be the last argument
		b := p.(*v2.BodyParameter)
		body := &pkg.Param{
			ID:   "request",
			Kind: pkg.Body,
			Type: tr.convertSchema(b.Schema, &pkg.TypeDecl{
				Name: methodName + "Request",
			}, false),
		}

		if t, ok := body.Type.(*pkg.IdentType); ok {
			if t.Name != methodName+"Request" {
				body.ID = formatVar(t.Name)
			}
		}
		body.Type = &pkg.PointerType{Type: body.Type}

		return body, nil, nil
	case "path":
		for i := range pathParams {
			if pathParams[i].ID != p.GetName() {
				continue
			}

			typ, _ := tr.typeForParameter(p)
			pathParams[i].ID = formatVar(pathParams[i].ID)
			pathParams[i].Type = typ
			break
		}

		return nil, nil, nil
	case "header", "query":
		k := pkg.Query
		if p.GetIn() == "header" {
			k = pkg.Header
		}

		typ, cf := tr.typeForParameter(p)

		if p.IsRequired() {
			param := pkg.Param{
				ID:         formatVar(p.GetName()),
				Orig:       p.GetName(),
				Type:       typ,
				Kind:       k,
				Collection: cf,
			}
			return nil, []pkg.Param{param}, nil
		}

		opt := pkg.Field{
			ID:         formatID(p.GetName()),
			Orig:       p.GetName(),
			Type:       &pkg.PointerType{Type: typ},
			Kind:       k,
			Collection: cf,
		}

		if p.GetDescription() != nil {
			opt.Comment = *p.GetDescription()
		}

		return nil, nil, []pkg.Field{opt}
	case "$ref":

	}

	return nil, nil, nil
}

// methodMap converts HTTP methods to preferred method name prefixes, based on
// which HTTP methods are supported on the given url.
// It does not handle conversion of Get to List.
func methodMap(methods map[string]*v2.Operation) map[string]string {
	out := make(map[string]string)
	for k := range methods {
		out[k] = k
	}

	_, hasPost := out["Post"]
	_, hasPut := out["Put"]
	_, hasPatch := out["Patch"]

	putMethod := "Create"
	if hasPost {
		out["Post"] = "Create"
		putMethod = "Replace"
	}

	if hasPatch {
		out["Patch"] = "Update"
		if hasPut {
			out["Put"] = putMethod
		}
	} else if hasPut {
		out["Put"] = "Update"
	}

	return out
}
