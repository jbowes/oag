package writer

import (
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/jbowes/oag/pkg"
)

func writeType(typ pkg.Type) func(s *jen.Statement) {
	return func(s *jen.Statement) {
		switch t := typ.(type) {
		case *pkg.IdentType:
			if t.Qualifier != "" {
				s.Qual(t.Qualifier, t.Name)
			} else {
				s.Id(t.Name)
			}
		case *pkg.SliceType:
			s.Index().Do(writeType(t.Type))
		case *pkg.PointerType:
			s.Op("*").Do(writeType(t.Type))
		case *pkg.IterType:
			s.Do(writeType(t.Type))
		case *pkg.StructType:
			s.Struct(convertFields(t.Fields)...)
		default:
			panic("unhandled type")
		}
	}
}

func convertFields(fields []pkg.Field) []jen.Code {
	var o []jen.Code
	first := true
	for _, f := range fields {
		sf := jen.Id(f.ID)
		sf.Do(writeType(f.Type))

		if f.Orig != "" {
			sf.Tag(map[string]string{"json": f.Orig})
		}

		blankAdded := false
		if !first && hasStruct(f.Type) {
			o = append(o, jen.Empty())
			blankAdded = true
		}

		if f.Comment != "" {
			if len(f.Comment) < 80 {
				sf.Comment(strings.Replace(f.Comment, "\n", " ", -1))
			} else {
				if !first && !blankAdded {
					o = append(o, jen.Empty())
				}
				o = append(o, jen.Comment(formatComment(f.Comment)))
			}
		}

		o = append(o, sf)
		if first {
			first = false
		}
	}

	return o
}

func hasStruct(typ pkg.Type) bool {
	switch t := typ.(type) {
	case *pkg.IterType:
		return hasStruct(t.Type)
	case *pkg.SliceType:
		return hasStruct(t.Type)
	case *pkg.PointerType:
		return hasStruct(t.Type)
	case *pkg.StructType:
		return true
	default:
		return false
	}
}

func typeName(typ pkg.Type) string {
	switch t := typ.(type) {
	case *pkg.IdentType:
		return t.Name
	case *pkg.IterType:
		return typeName(t.Type)
	case *pkg.SliceType:
		return typeName(t.Type)
	case *pkg.PointerType:
		return typeName(t.Type)
	default:
		panic("unhandled type")
	}
}

func initType(typ pkg.Type) jen.Code {
	ret := jen.Empty()

	if t, ok := typ.(*pkg.PointerType); ok {
		ret = jen.Op("&")
		typ = t.Type
	}

	ret.Do(writeType(typ)).Block()
	return ret
}
