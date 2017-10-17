package translator

import (
	"fmt"
	"strings"

	"github.com/jbowes/oag/openapi/v2"
	"github.com/jbowes/oag/pkg"
)

type typeRegistry struct {
	strFmt stringFormat
	types  []pkg.TypeDecl
}

func (tr *typeRegistry) add(td pkg.TypeDecl) {
	tr.types = append(tr.types, td)
}

func (tr *typeRegistry) convertSchema(schema v2.Schema, td *pkg.TypeDecl, declAll bool) pkg.Type {
	var ret pkg.Type
	switch s := schema.(type) {
	case *v2.ObjectSchema:
		t := &pkg.StructType{}
		if s.Properties == nil {
			if !s.AnyAdditionalProperties && s.AdditionalProperties == nil {
				// no properties at all. empty struct.
				return t
			}

			mt := &pkg.MapType{Key: &pkg.IdentType{Name: "string"}}

			if s.AnyAdditionalProperties {
				mt.Value = &pkg.InterfaceType{}
			} else {
				mt.Value = tr.convertSchema(s.AdditionalProperties, &pkg.TypeDecl{
					Name: td.Name + "Value"}, false)
			}

			if td != nil {
				td.Type = mt
				tr.types = append(tr.types, *td)
			}

			return &pkg.IdentType{Name: td.Name}
		}

		required := make(map[string]struct{})
		if s.Required != nil {
			for _, field := range *s.Required {
				required[field] = struct{}{}
			}
		}

		for _, prop := range *s.Properties {
			field := pkg.Field{
				ID: formatID(prop.Name),
			}

			if field.ID != prop.Name {
				field.Orig = prop.Name
			}

			fieldComment := ""
			if prop.Schema.GetTitle() != nil {
				fieldComment += *prop.Schema.GetTitle()
			}
			if prop.Schema.GetDescription() != nil {
				fieldComment += *prop.Schema.GetDescription()
			}
			field.Comment = fieldComment

			sn := td.Name + formatID(prop.Name)
			field.Type = tr.convertSchema(prop.Schema, &pkg.TypeDecl{
				Name:    sn,
				Comment: fmt.Sprintf("%s is a data type for API communication.", sn),
			}, false)

			if _, ok := required[prop.Name]; !ok {
				field.Type = &pkg.PointerType{Type: field.Type}
				if field.Comment == "" {
					field.Comment = "Optional"
				}
			}

			t.Fields = append(t.Fields, field)
		}
		if td != nil {
			td.Type = t
			tr.types = append(tr.types, *td)
		}

		return &pkg.IdentType{Name: td.Name}
	case *v2.StringSchema:
		ret = tr.strFmt.typeFor(s.Format)
	case *v2.IntegerSchema:
		ret = &pkg.IdentType{Name: "int"}
	case *v2.NumberSchema:
		ret = &pkg.IdentType{Name: "float64"}
	case *v2.BooleanSchema:
		ret = &pkg.IdentType{Name: "bool"}
	case *v2.ReferenceSchema:
		parts := strings.Split(s.Reference, "/")
		ret = &pkg.IdentType{Name: formatID(parts[len(parts)-1])}
	case *v2.ArraySchema:
		ret = &pkg.SliceType{Type: tr.convertSchema(s.Items, &pkg.TypeDecl{
			Name: td.Name + "Items",
		}, false)}
	default:
		// XXX handle this
		panic("unknown type")
	}

	if td != nil && declAll {
		td.Type = ret
		tr.types = append(tr.types, *td)
	}

	return ret
}

func (tr *typeRegistry) typeForParameter(p v2.Parameter) (pkg.Type, pkg.Collection) {
	switch t := p.(type) {
	case *v2.StringParameter:
		return tr.strFmt.typeFor(t.Format), pkg.None
	case *v2.NumberParameter:
		return &pkg.IdentType{Name: "float64"}, pkg.None
	case *v2.IntegerParameter:
		return &pkg.IdentType{Name: "int"}, pkg.None
	case *v2.BooleanParameter:
		return &pkg.IdentType{Name: "bool"}, pkg.None
	case *v2.ArrayParameter:
		cf := "csv" // the default
		if t.CollectionFormat != nil {
			cf = *t.CollectionFormat
		}

		var collection pkg.Collection
		switch cf {
		case "csv":
			collection = pkg.CSV
		case "ssv":
			collection = pkg.SSV
		case "tsv":
			collection = pkg.TSV
		case "pipes":
			collection = pkg.Pipes
		case "multi":
			collection = pkg.Multi
		default:
			panic("unsupported collection format")
		}
		return &pkg.SliceType{Type: tr.convertItems(t.Items)}, collection
	default:
		panic("unhandled parameter type")
	}
}

func (tr *typeRegistry) convertItems(i v2.Items) pkg.Type {
	switch t := i.(type) {
	case *v2.StringItem:
		return tr.strFmt.typeFor(t.Format)
	case *v2.NumberItem:
		return &pkg.IdentType{Name: "float64"}
	case *v2.IntegerItem:
		return &pkg.IdentType{Name: "int"}
	case *v2.BooleanItem:
		return &pkg.IdentType{Name: "bool"}
	case *v2.ArrayItem:
		return &pkg.SliceType{Type: tr.convertItems(t.Items)}
	default:
		panic("unhandled item type")
	}
}

type stringFormat map[string]string

func (sf stringFormat) typeFor(fmt *string) pkg.Type {
	if fmt == nil {
		return &pkg.IdentType{Name: "string"}
	}

	if f, ok := sf[*fmt]; ok {

		idx := strings.LastIndex(f, ".")
		return &pkg.IdentType{
			Name:      f[idx+1:],
			Qualifier: f[:idx],
			Marshal:   true,
		}
	}

	return &pkg.IdentType{Name: "string"}
}
