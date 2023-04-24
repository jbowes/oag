package translator

import (
	"fmt"
	"strings"

	"github.com/jbowes/oag/openapi/v2"
	"github.com/jbowes/oag/pkg"
)

type typeRegistry struct {
	strFmt         stringFormat
	types          []pkg.TypeDecl
	discriminators map[string]*pkg.TypeDecl
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
				mt.Value = &pkg.EmptyInterfaceType{}
			} else {
				mt.Value = tr.convertSchema(s.AdditionalProperties, &pkg.TypeDecl{
					Name: td.Name + "Value"}, false)
			}

			td.Type = mt
			tr.types = append(tr.types, *td)
			return &pkg.IdentType{Name: td.Name}
		}

		required := make(map[string]struct{})
		if s.Required != nil {
			for _, field := range *s.Required {
				required[field] = struct{}{}
			}
		}

		if s.Discriminator != nil {
			t := &pkg.InterfaceType{}

			for _, prop := range *s.Properties {
				method := pkg.InterfaceMethod{
					Name: formatID("Get", prop.Name),
				}

				method.Comment = commentForPropSchema(prop.Schema)

				sn := td.Name + formatID(prop.Name)
				method.Return = tr.convertSchema(prop.Schema, &pkg.TypeDecl{
					Name:    sn,
					Comment: fmt.Sprintf("%s is a data type for API communication.", sn),
				}, false)

				if _, ok := required[prop.Name]; !ok {
					method.Return = &pkg.PointerType{Type: method.Return}
					if method.Comment == "" {
						method.Comment = "Optional"
					}
				}

				t.Methods = append(t.Methods, method)
			}

			td.Type = t
			tr.types = append(tr.types, *td)

			if tr.discriminators == nil {
				tr.discriminators = make(map[string]*pkg.TypeDecl)
			}
			tr.discriminators[td.Name] = &tr.types[len(tr.types)-1]

			oldDiscriminator := s.Discriminator
			s.Discriminator = nil
			mn := td.Name + "Meta"
			tr.convertSchema(s, &pkg.TypeDecl{
				Name:    mn,
				Comment: fmt.Sprintf("%s is an abstract data type for API communication.", mn),
			}, false)

			s.Discriminator = oldDiscriminator
			return &pkg.IdentType{Name: td.Name}
		}

		for _, prop := range *s.Properties {
			field := pkg.Field{
				ID: formatID(prop.Name),
			}

			if field.ID != prop.Name {
				field.Orig = prop.Name
			}

			field.Comment = commentForPropSchema(prop.Schema)

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

		td.Type = t
		tr.types = append(tr.types, *td)
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
	case *v2.AllOfSchema:
		fields := make([]pkg.Field, len(s.AllOf))

		for i := range s.AllOf {
			if dr, ok := s.AllOf[i].(*v2.ReferenceSchema); ok {
				parts := strings.Split(dr.Reference, "/")
				refName := formatID(parts[len(parts)-1])

				if dt, ok := tr.discriminators[refName]; ok {
					fields[i] = pkg.Field{
						Type: &pkg.IdentType{
							Name: refName + "Meta",
						},
					}

					td.Comment = strings.Replace(td.Comment, td.Name, td.Name+refName, 1)
					td.Name += refName

					disc := dt.Type.(*pkg.InterfaceType)
					disc.Implementors = append(disc.Implementors, pkg.InterfaceImplementor{
						Discriminator: td.Orig,
						Type:          &pkg.IdentType{Name: td.Name},
					})

					continue
				}
			}

			sn := fmt.Sprintf("%sAllOf%d", td.Name, i)
			field := pkg.Field{
				Type: tr.convertSchema(s.AllOf[i], &pkg.TypeDecl{
					Name:    sn,
					Comment: fmt.Sprintf("%s is a data type for API communication.", sn),
				}, false),
			}

			fields[i] = field
		}

		t := &pkg.StructType{
			Fields: fields,
		}

		td.Type = t
		tr.types = append(tr.types, *td)
		return &pkg.IdentType{Name: td.Name}
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

// indirect wraps a type in a pointer for use in parameters / return values,
// if required.
func (tr *typeRegistry) indirect(t pkg.Type) pkg.Type {
	return &pkg.PointerType{Type: t}
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

func commentForPropSchema(prop v2.Schema) string {
	comment := ""
	if prop.GetTitle() != nil {
		comment += *prop.GetTitle()
	}
	if prop.GetDescription() != nil {
		comment += *prop.GetDescription()
	}

	return comment
}
