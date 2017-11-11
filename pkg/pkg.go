// Package pkg contains the structures that define a generated go client api,
package pkg

// Package is a go package
type Package struct {
	Qualifier string
	Name      string

	BaseURL string

	TypeDecls []TypeDecl

	Iters   []Iter
	Clients []Client
}

// Type is type literal or qualified identifier. It may be used inline, or
// as part of a type declaration.
type Type interface {
	Equal(Type) bool
}

// IdentType is a qualified identifier for another type
type IdentType struct {
	Qualifier string
	Name      string
	Marshal   bool // does this use marshaltext or is it natively represented
}

// Equal implements equality for Types
func (t *IdentType) Equal(o Type) bool {
	if ot, ok := o.(*IdentType); ok {
		return *t == *ot
	}

	return false
}

// PointerType is a pointer type
type PointerType struct {
	Type
}

// Equal implements equality for Types
func (t *PointerType) Equal(o Type) bool {
	if ot, ok := o.(*PointerType); ok {
		return t.Type.Equal(ot.Type)
	}

	return false
}

// SliceType is a slice type
type SliceType struct {
	Type
}

// Equal implements equality for Types
func (t *SliceType) Equal(o Type) bool {
	if ot, ok := o.(*SliceType); ok {
		return t.Type.Equal(ot.Type)
	}

	return false
}

// StructType is a struct type
type StructType struct {
	Fields []Field
}

// Equal implements equality for Types
func (t *StructType) Equal(o Type) bool {
	if ot, ok := o.(*StructType); ok {
		if len(t.Fields) != len(ot.Fields) {
			return false
		}

		for i, f := range t.Fields {
			if !f.equal(ot.Fields[i]) {
				return false
			}
		}

		return true
	}

	return false
}

// IterType is used for return types, indicating they're iterators
type IterType struct {
	Type
}

// Equal implements equality for Types
func (t *IterType) Equal(o Type) bool {
	if ot, ok := o.(*IterType); ok {
		return t.Type.Equal(ot.Type)
	}

	return false
}

// MapType is a map type
type MapType struct {
	Key   Type
	Value Type
}

// Equal implements equality for Types
func (t *MapType) Equal(o Type) bool {
	if ot, ok := o.(*MapType); ok {
		return t.Key.Equal(ot.Key) && t.Value.Equal(ot.Value)
	}

	return false
}

// EmptyInterfaceType is an empty interface
type EmptyInterfaceType struct{}

// Equal implements equality for Types
func (t *EmptyInterfaceType) Equal(o Type) bool {
	_, ok := o.(*EmptyInterfaceType)
	return ok
}

// TypeDecl is a type declaration.
type TypeDecl struct {
	Name    string
	Comment string
	Type    Type
}

// Iter is an iterator over multiple results/pages from a response
type Iter struct {
	Name   string
	Return Type
}

// Field is a struct field
type Field struct {
	ID      string
	Type    Type
	Comment string

	Orig       string     // optional name of field as it is originally from the spec
	Kind       Kind       // optional. Used for Opts structs
	Collection Collection // optional. Used for Opts structs
}

func (f Field) equal(of Field) bool {
	return f.ID == of.ID &&
		f.Type.Equal(of.Type) &&
		f.Comment == of.Comment &&
		f.Orig == of.Orig &&
		f.Kind == of.Kind &&
		f.Collection == of.Collection
}

// Client is a struct that holds the methods for communicating with an API
// endpoint
type Client struct {
	Name        string
	Comment     string
	ContextName string

	Methods []Method
}

// Method is a struct method on a Client for calling a remote API
type Method struct {
	Receiver struct {
		ID   string
		Arg  string
		Type string
	}

	Name    string
	Params  []Param
	Return  []Type
	Comment string

	Errors map[int]Type // Non-success status codes to types. -1 is default

	HTTPMethod string
	Path       string // Path to endpoint, in printf format, including base path.
}

// Kind is the kind of parameter; ie where it maps to in the request
type Kind uint8

// The possible Kinds
const (
	Body Kind = iota
	Query
	Path
	Header

	Opts // Opts struct holding optional values
)

// Collection specifies how to encode slice parameters
type Collection uint8

// The possible Collection formats
const (
	None Collection = 0
	CSV  Collection = iota
	SSV
	TSV
	Pipes
	Multi
)

// Param is a function parameter
type Param struct {
	ID         string
	Orig       string // original name, ie for query params or headers
	Arg        string // argument name, this is to avoid reserved keywords being used
	Type       Type
	Kind       Kind
	Collection Collection
}
