// Package v2 defines the OpenAPI 2.0 specification data structures.
//
// The specification can be found at
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md
package v2

import (
	"errors"
	"net/url"

	"github.com/go-yaml/yaml"
)

// Document is a top level OpenAPI 2.0 / Swagger 2.0 API definition, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#swagger-object
type Document struct {
	Version string // required
	Info    *Info  // required

	Host     *string
	BasePath *string `yaml:"basePath"`
	Schemes  []string
	Consumes []string
	Produces []string

	Paths map[string]PathItem // required

	Definitions *SchemaMap
	Parameters  *ParameterMap
	Responses   *map[string]Response

	SecurityDefinitions *SecuritySchemeMap `yaml:"securityDefinitions"`
	Security            []map[string][]string

	Tags          []Tag
	Documentation *ExternalDocumentation
}

// Info is the required OpenAPI Info object, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#infoObject
type Info struct {
	Title       string // required
	Description *string

	TermsOfService *string
	Contact        *Contact
	License        *License

	Version string // required
}

// Contact holds the API contact information, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#contactObject
type Contact struct {
	Name  *string
	URL   *url.URL
	Email *string
}

// License holds the API license information, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#licenseObject
type License struct {
	Name string // required
	URL  *url.URL
}

// PathItem describes all operations/methods at a single path, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#pathItemObject
type PathItem struct {
	Reference *string

	Get     *Operation
	Put     *Operation
	Post    *Operation
	Delete  *Operation
	Options *Operation
	Head    *Operation
	Patch   *Operation

	Parameters Parameters
}

// Operation describes a single method on a given path, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#operationObject
type Operation struct {
	Tags []string

	Summary       *string
	Description   *string
	Documentation *ExternalDocumentation

	OperationID *string

	Consumes []string
	Produces []string

	Parameters Parameters
	Responses  *Responses // required

	Schemes []string

	Deprecated bool

	Security []map[string][]string
}

// ExternalDocumentation describes a reference to additional documentation hosted
// elsewhere, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#externalDocumentationObject
type ExternalDocumentation struct {
	Description *string
	URL         *url.URL // required
}

// Parameters is a slice of Parameter structs.
type Parameters []Parameter

// UnmarshalYAML unmarshals Parameters from YAML or JSON.
func (p *Parameters) UnmarshalYAML(um func(interface{}) error) error {
	var ys []map[string]interface{}
	if err := um(&ys); err != nil {
		return err
	}

	*p = make(Parameters, len(ys))

	var err error
	for i, y := range ys {
		if (*p)[i], err = unmarshalParameter(y); err != nil {
			return err
		}
	}

	return nil
}

// ParameterMap is a map of identifiers to Parameters
type ParameterMap map[string]Parameter

// UnmarshalYAML unmarshals a ParameterMap from YAML or JSON.
func (p *ParameterMap) UnmarshalYAML(um func(interface{}) error) error {
	var ys map[string]map[string]interface{}
	if err := um(&ys); err != nil {
		return err
	}
	*p = make(map[string]Parameter, len(ys))

	var err error
	for k, y := range ys {
		if (*p)[k], err = unmarshalParameter(y); err != nil {
			return err
		}
	}

	return nil
}

func unmarshalParameter(y map[string]interface{}) (Parameter, error) {
	b, err := yaml.Marshal(&y)
	if err != nil {
		return nil, err
	}

	if _, ok := y["$ref"]; ok {
		var v ReferenceParamter
		err = yaml.Unmarshal(b, &v)
		return v, err
	} else if y["in"] == "body" {
		var v BodyParameter
		err = yaml.Unmarshal(b, &v)
		return &v, err
	}

	switch y["type"] {
	case "string":
		var v StringParameter
		err = yaml.Unmarshal(b, &v)
		return &v, err
	case "number":
		var v NumberParameter
		err = yaml.Unmarshal(b, &v)
		return &v, err
	case "integer":
		var v IntegerParameter
		err = yaml.Unmarshal(b, &v)
		return &v, err
	case "boolean":
		var v BooleanParameter
		err = yaml.Unmarshal(b, &v)
		return &v, err
	case "array":
		// XXX for whatever reason, ParameterFields won't unmarshal in
		// ArrayParameter directly.
		var p ParameterFields
		if err = yaml.Unmarshal(b, &p); err != nil {
			return nil, err
		}
		v := ArrayParameter{ParameterFields: p}
		err = yaml.Unmarshal(b, &v)
		return &v, err
	case "file":
		var v FileParameter
		err = yaml.Unmarshal(b, &v)
		return &v, err
	default:
		return nil, errors.New("bad parameter type")
	}
}

// Parameter is the common interface for types that define operation parameters,
// according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#parameterObject
type Parameter interface {
	GetName() string // required
	GetIn() string   // required
	GetDescription() *string
	IsRequired() bool
}

// ReferenceParamter is a $ref to a parameter defined in the parameters section
// of the OpenAPI doc.
type ReferenceParamter struct {
	Reference string `yaml:"$ref"`
}

// GetName returns the name of this parameter.
func (ReferenceParamter) GetName() string { return "" }

// GetIn returns the location (path, query, header, body) of this parameter.
func (ReferenceParamter) GetIn() string { return "" }

// GetDescription returns the optional documentation description for this parameter.
func (ReferenceParamter) GetDescription() *string { return nil }

// IsRequired returns if this parameter is required for the operation.
func (ReferenceParamter) IsRequired() bool { return false }

// ParameterFields holds the common fields for all non-reference parameter types
type ParameterFields struct {
	Name        string // required
	In          string // required
	Description *string
	Required    bool
}

// GetName returns the name of this parameter.
func (pf *ParameterFields) GetName() string { return pf.Name }

// GetIn returns the location (path, query, header, body) of this parameter.
func (pf *ParameterFields) GetIn() string { return pf.In }

// GetDescription returns the description field of the parameter.
func (pf *ParameterFields) GetDescription() *string { return pf.Description }

// IsRequired returns if this parameter is required for the operation.
func (pf *ParameterFields) IsRequired() bool { return pf.Required }

// BodyParameter is an operation parameter that represents a request body.
type BodyParameter struct {
	ParameterFields `yaml:",inline"`
	Schema          Schema // required
}

// UnmarshalYAML unmarshals a BodyParameter from YAML or JSON.
func (b *BodyParameter) UnmarshalYAML(um func(interface{}) error) error {
	var y struct {
		ParameterFields `yaml:",inline"`
		Schema          yaml.MapSlice
	}

	if err := um(&y); err != nil {
		return err
	}

	b.ParameterFields = y.ParameterFields

	var err error
	b.Schema, err = unmarshalSchema(y.Schema)
	return err
}

// StringParameter is an operation parameter in any non-body location that is a
// string type.
type StringParameter struct {
	ParameterFields `yaml:",inline"`
	StringItem      `yaml:",inline"`
	AllowEmptyValue bool
}

// NumberParameter is an operation parameter in any non-body location that is a
// number type.
type NumberParameter struct {
	ParameterFields `yaml:",inline"`
	NumberItem      `yaml:",inline"`
	AllowEmptyValue bool
}

// IntegerParameter is an operation parameter in any non-body location that is
// an integer type.
type IntegerParameter struct {
	ParameterFields `yaml:",inline"`
	IntegerItem     `yaml:",inline"`
	AllowEmptyValue bool
}

// BooleanParameter is an operation parameter in any non-body location that is a
// boolean type.
type BooleanParameter struct {
	ParameterFields `yaml:",inline"`
	BooleanItem     `yaml:",inline"`
	AllowEmptyValue bool
}

// ArrayParameter is an operation parameter in any non-body location that is an
// array type.
type ArrayParameter struct {
	ParameterFields `yaml:",inline"`
	ArrayItem       `yaml:",inline"`
	AllowEmptyValue bool
}

// FileParameter is an operation parameter that represents a file upload.
type FileParameter struct {
	ParameterFields `yaml:",inline"`
	AllowEmptyValue bool
}

// Items are used for array item definitions, and parameters.

// Items is the common interface for types that define array item, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#itemsObject
type Items interface {
	Type() string // required
}

const (
	stringItem  = "string"
	numberItem  = "number"
	integerItem = "integer"
	booleanItem = "boolean"
	arrayItem   = "array"
)

func unmarshalItems(y map[string]interface{}) (Items, error) {
	b, err := yaml.Marshal(&y)
	if err != nil {
		return nil, err
	}

	switch y["type"] {
	case stringItem:
		var v StringItem
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case numberItem:
		var v NumberItem
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case integerItem:
		var v IntegerItem
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case booleanItem:
		var v BooleanItem
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case arrayItem:
		var v ArrayItem
		err := yaml.Unmarshal(b, &v)
		return &v, err
	default:
		return nil, errors.New("bad schema type: " + y["type"].(string))
	}
}

// StringItem represents the definition for a string array item.
type StringItem struct {
	Format *string

	Default *string
	Enum    *[]string

	MaxLength *int64
	MinLength *int64
	Pattern   *string
}

// Type returns the type of this item.
func (StringItem) Type() string { return stringItem }

// NumberItem represents the definition for a number array item.
type NumberItem struct {
	Format *string

	Default *float64
	Enum    *[]float64

	Maximum          *float64
	ExclusiveMaximum bool
	Minium           *float64
	ExclusiveMinimum bool
	MultipleOf       *float64
}

// Type returns the type of this item.
func (NumberItem) Type() string { return numberItem }

// IntegerItem represents the definition for an integer array item.
type IntegerItem struct {
	Format *string

	Default *int64
	Enum    *[]int64

	Maximum          *int64
	ExclusiveMaximum bool
	Minium           *int64
	ExclusiveMinimum bool
	MultipleOf       *int64
}

// Type returns the type of this item.
func (IntegerItem) Type() string { return integerItem }

// BooleanItem represents the definition for a boolean array item.
type BooleanItem struct {
	Default *bool
	Enum    *[]bool
}

// Type returns the type of this item.
func (BooleanItem) Type() string { return booleanItem }

// ArrayFields defines common fields for array definitions.
type ArrayFields struct {
	CollectionFormat *string `yaml:"collectionFormat"`

	Default interface{}
	Enum    *[]interface{}

	MaxItems    *uint64
	MinItems    *uint64
	UniqueItems bool
}

// ArrayItem represents the definition for a nested array array item.
type ArrayItem struct {
	ArrayFields `yaml:",inline"`
	Items       Items // required
}

// Type returns the type of this item.
func (ArrayItem) Type() string { return arrayItem }

// UnmarshalYAML unmarshals an ArrayItem from YAML or JSON.
func (a *ArrayItem) UnmarshalYAML(um func(interface{}) error) error {
	var y struct {
		ArrayFields `yaml:",inline"`

		Items map[string]interface{}
	}

	if err := um(&y); err != nil {
		return err
	}

	a.ArrayFields = y.ArrayFields

	var err error
	a.Items, err = unmarshalItems(y.Items)
	return err
}

// Responses defines the Response for each status code for an operation,
// according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responsesObject
type Responses struct {
	Default *Response
	Codes   map[int]Response
}

// UnmarshalYAML unmarshals Responses from YAML or JSON.
func (r *Responses) UnmarshalYAML(um func(interface{}) error) error {
	var y map[interface{}]Response
	if err := um(&y); err != nil {
		return err
	}

	for k, v := range y {
		switch t := k.(type) {
		case int:
			if r.Codes == nil {
				r.Codes = make(map[int]Response)
			}
			r.Codes[t] = v
		case string:
			nv := v
			r.Default = &nv
		}
	}

	return nil
}

// Response is a single resposne from an operation, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responseObject
//
// XXX handle reference
type Response struct {
	Description string // required
	Schema      Schema
	Headers     map[string]Header
	Examples    map[string]interface{} // mime type to anything
}

// UnmarshalYAML unmarshals a Response from YAML or JSON.
func (r *Response) UnmarshalYAML(um func(interface{}) error) error {
	var y struct {
		Description string
		Headers     map[string]Header
		Examples    map[string]interface{}

		Schema yaml.MapSlice
	}

	if err := um(&y); err != nil {
		return err
	}

	r.Description = y.Description
	r.Headers = y.Headers
	r.Examples = y.Examples

	if y.Schema == nil { // schema is optional for responses. they may have no body
		return nil
	}

	var err error
	r.Schema, err = unmarshalSchema(y.Schema)
	return err
}

// Header is the common interface for headers set on responses, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#headerObject
type Header interface {
	Description() *string
	Items
}

// HeaderFields holds the common fields for all response header types.
type HeaderFields struct {
	Description *string
}

// GetDescription returns the optional documentation description for this header.
func (h *HeaderFields) GetDescription() *string { return h.Description }

// StringHeader is a string type response header.
type StringHeader struct {
	HeaderFields `yaml:",inline"`
	StringItem   `yaml:",inline"`
}

// NumberHeader is a number type response header.
type NumberHeader struct {
	HeaderFields `yaml:",inline"`
	NumberItem   `yaml:",inline"`
}

// IntegerHeader is an integer type response header.
type IntegerHeader struct {
	HeaderFields `yaml:",inline"`
	IntegerItem  `yaml:",inline"`
}

// BooleanHeader is a boolean type response header.
type BooleanHeader struct {
	HeaderFields `yaml:",inline"`
	BooleanItem  `yaml:",inline"`
}

// ArrayHeader is an array type response header.
type ArrayHeader struct {
	HeaderFields `yaml:",inline"`
	ArrayItem    `yaml:",inline"`
}

// Tag defines additional metadata for tags attached to operations, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#tag-object
type Tag struct {
	Name          string // required
	Description   *string
	Documentation *ExternalDocumentation
}

// Schema defines the common interface for schema definitions, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#schemaObject
type Schema interface {
	GetTitle() *string
	GetDescription() *string
	GetDocumentation() *ExternalDocumentation
	GetExample() interface{}
}

//SchemaMap is an ordered list of named schema definitions or object properties,
// so that order is maintained.
type SchemaMap []struct {
	Name   string
	Schema Schema
}

// UnmarshalYAML unmarshals a SchemaMap from YAML or JSON.
func (s *SchemaMap) UnmarshalYAML(um func(interface{}) error) error {
	var ys yaml.MapSlice
	if err := um(&ys); err != nil {
		return err
	}
	*s = make(SchemaMap, len(ys))

	for i, y := range ys {
		v, err := unmarshalSchema(y.Value.(yaml.MapSlice))
		if err != nil {
			return err
		}
		(*s)[i] = struct {
			Name   string
			Schema Schema
		}{y.Key.(string), v}
	}

	return nil
}

func unmarshalSchema(yms yaml.MapSlice) (Schema, error) {
	b, err := yaml.Marshal(&yms)
	if err != nil {
		return nil, err
	}

	// XXX wasteful just for indexing
	var y map[string]interface{}
	if err = yaml.Unmarshal(b, &y); err != nil {
		return nil, err
	}

	if _, ok := y["$ref"]; ok {
		var v ReferenceSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	} else if _, ok := y["allOf"]; ok {
		var v AllOfSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	}

	switch y["type"] {
	case "object":
		var v ObjectSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "null":
		var v NullSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "string":
		var v StringSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "number":
		var v NumberSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "integer":
		var v IntegerSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "boolean":
		var v BooleanSchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "array":
		var v ArraySchema
		err := yaml.Unmarshal(b, &v)
		return &v, err
	default:
		return nil, errors.New("bad schema type: " + y["type"].(string))
	}
}

// ReferenceSchema is an inline reference to another schema definition.
type ReferenceSchema struct {
	Reference string `yaml:"$ref"` // required
}

// GetTitle returns the optional title for this schema.
func (ReferenceSchema) GetTitle() *string { return nil }

// GetDescription returns the optional documentation description for this schema.
func (ReferenceSchema) GetDescription() *string { return nil }

// GetDocumentation returns the optional Documentation for this schema.
func (ReferenceSchema) GetDocumentation() *ExternalDocumentation { return nil }

// GetExample returns the optional example value for this schema.
func (ReferenceSchema) GetExample() interface{} { return nil }

// SchemaFields holds the common fields for schema definitions.
type SchemaFields struct {
	Title         *string
	Description   *string
	Documentation *ExternalDocumentation
	Example       interface{}

	ReadOnly bool        // valid only for items under properties
	XML      interface{} // valid only for items under properties
}

// GetTitle returns the optional title for this schema.
func (s *SchemaFields) GetTitle() *string { return s.Title }

// GetDescription returns the optional documentation description for this schema.
func (s *SchemaFields) GetDescription() *string { return s.Description }

// GetDocumentation returns the optional Documentation for this schema.
func (s *SchemaFields) GetDocumentation() *ExternalDocumentation { return s.Documentation }

// GetExample returns the optional example value for this schema.
func (s *SchemaFields) GetExample() interface{} { return s.Example }

// AllOfSchema represents an allOf definition, according to
// https://tools.ietf.org/html/draft-fge-json-schema-validation-00#section-5.5.3
type AllOfSchema struct {
	SchemaFields `yaml:",inline"`
	AllOf        []Schema `yaml:"allOf"`
}

// UnmarshalYAML unmarshals an AllOfSchema from YAML or JSON.
func (a *AllOfSchema) UnmarshalYAML(um func(interface{}) error) error {
	var ay struct {
		SchemaFields `yaml:",inline"`
		AllOf        []yaml.MapSlice `yaml:"allOf"`
	}
	if err := um(&ay); err != nil {
		return err
	}
	a.AllOf = make([]Schema, len(ay.AllOf))

	var err error
	for i, y := range ay.AllOf {
		if a.AllOf[i], err = unmarshalSchema(y); err != nil {
			return err
		}
	}

	return nil
}

// ObjectSchema is a schema definition for an object.
type ObjectSchema struct {
	SchemaFields  `yaml:",inline"`
	Descriminator *string

	Properties *SchemaMap
	Required   *[]string

	MinProperties *uint64
	MaxProperties *uint64
}

// NullSchema is a literal null schema definition.
type NullSchema struct {
	SchemaFields `yaml:",inline"`
}

// StringSchema is a schema definition for a string type.
type StringSchema struct {
	SchemaFields `yaml:",inline"`
	StringItem   `yaml:",inline"`
}

// NumberSchema is a schema definition for a number type.
type NumberSchema struct {
	SchemaFields `yaml:",inline"`
	NumberItem   `yaml:",inline"`
}

// IntegerSchema is a schema definition for an integer type.
type IntegerSchema struct {
	SchemaFields `yaml:",inline"`
	IntegerItem  `yaml:",inline"`
}

// BooleanSchema is a schema definition for a boolean type.
type BooleanSchema struct {
	SchemaFields `yaml:",inline"`
	BooleanItem  `yaml:",inline"`
}

// ArraySchema is a schema definition for an array type.
type ArraySchema struct {
	SchemaFields `yaml:",inline"`
	ArrayFields  `yaml:",inline"`
	Items        Schema // required
}

// UnmarshalYAML unmarshals an ArraySchema from YAML or JSON.
func (a *ArraySchema) UnmarshalYAML(um func(interface{}) error) error {
	var y struct {
		SchemaFields `yaml:",inline"`
		ArrayFields  `yaml:",inline"`
		Items        yaml.MapSlice
	}

	if err := um(&y); err != nil {
		return err
	}

	a.SchemaFields = y.SchemaFields
	a.ArrayFields = y.ArrayFields

	var err error
	a.Items, err = unmarshalSchema(y.Items)
	return err
}

// SecuritySchemeMap is a map if identifiers to SecuritySchemes.
type SecuritySchemeMap map[string]SecurityScheme

// UnmarshalYAML unmarshals a SecuritySchemaMap from YAML or JSON.
func (s *SecuritySchemeMap) UnmarshalYAML(um func(interface{}) error) error {
	var ys map[string]map[string]interface{}
	if err := um(&ys); err != nil {
		return err
	}

	*s = make(map[string]SecurityScheme, len(ys))

	var err error
	for k, y := range ys {
		if (*s)[k], err = unmarshalSecurityScheme(y); err != nil {
			return err
		}
	}

	return nil
}

// SecurityScheme is the common interface for security schemes, according to
// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#securitySchemeObject
type SecurityScheme interface {
	Type() string // required
	GetDescription() *string
}

func unmarshalSecurityScheme(y map[string]interface{}) (SecurityScheme, error) {
	b, err := yaml.Marshal(&y)
	if err != nil {
		return nil, err
	}

	switch y["type"] {
	case "basic":
		var v BasicSecurityScheme
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "apiKey":
		var v APIKeySecurityScheme
		err := yaml.Unmarshal(b, &v)
		return &v, err
	case "oauth2":
		var v OAuth2SecurityScheme
		err := yaml.Unmarshal(b, &v)
		return &v, err
	default:
		return nil, errors.New("bad security scheme: " + y["type"].(string))
	}
}

// SecuritySchemeFields holds the common fields for SecuritySchemes.
type SecuritySchemeFields struct {
	Description *string
}

// GetDescription returns the optional documentation description for this security scheme.
func (s *SecuritySchemeFields) GetDescription() *string { return s.Description }

// BasicSecurityScheme represents a basic auth security scheme.
type BasicSecurityScheme struct {
	SecuritySchemeFields `yaml:",inline"`
}

// Type returns the type of this SecurityScheme
func (BasicSecurityScheme) Type() string { return "basic" }

// APIKeySecurityScheme represents an API key security scheme, sent via a header
// or query parameter.
type APIKeySecurityScheme struct {
	SecuritySchemeFields `yaml:",inline"`
	Name                 string // required
	In                   string // required
}

// Type returns the type of this SecurityScheme
func (APIKeySecurityScheme) Type() string { return "apiKey" }

// OAuth2SecurityScheme represents an OAuth 2.0 security scheme.
type OAuth2SecurityScheme struct {
	SecuritySchemeFields `yaml:",inline"`
	Flow                 string // required

	AuthorizationURL *url.URL `yaml:"authorizationUrl"` // required for implicit or accessCode flows
	TokenURL         *url.URL `yaml:"tokenUrl"`         // required for password, application, and accessCode flows.

	Scopes map[string]string // required
}

// UnmarshalYAML unmarshals an OAuth2SecurityScheme from YAML or JSON.
func (o *OAuth2SecurityScheme) UnmarshalYAML(um func(interface{}) error) error {
	var y struct {
		SecuritySchemeFields `yaml:",inline"`
		Flow                 string
		Scopes               map[string]string
		AuthorizationURL     *string `yaml:"authorizationUrl"`
		TokenURL             *string `yaml:"tokenUrl"`
	}
	if err := um(&y); err != nil {
		return err
	}

	o.SecuritySchemeFields = y.SecuritySchemeFields
	o.Flow = y.Flow
	o.Scopes = y.Scopes

	if y.AuthorizationURL != nil {
		u, err := url.Parse(*y.AuthorizationURL)
		if err != nil {
			return err
		}
		o.AuthorizationURL = u
	}

	if y.TokenURL != nil {
		u, err := url.Parse(*y.TokenURL)
		if err != nil {
			return err
		}
		o.TokenURL = u
	}

	return nil
}

// Type returns the type of this SecurityScheme
func (OAuth2SecurityScheme) Type() string { return "oauth2" }
