package v2

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/renstrom/dedent"
)

func TestParametersUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
	- name: param
	  in: query
	  type: string
	`)

	var out Parameters
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	expected := Parameters{
		&StringParameter{ParameterFields: ParameterFields{In: "query", Name: "param"}},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Error("Wrong value unmarshaled. got:", out, "expected:", expected)
	}
}

func TestParameterMapUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
	qparam:
	  name: param
	  in: query
	  type: string
	`)

	var out ParameterMap
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	expected := ParameterMap{
		"qparam": &StringParameter{ParameterFields: ParameterFields{In: "query", Name: "param"}},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Error("Wrong value unmarshaled. got:", out, "expected:", expected)
	}
}

func TestUnmarshalParameter(t *testing.T) {
	tcs := []struct {
		name string
		in   string
		out  Parameter
	}{
		{
			"body",
			`
            - name: body
              in: body
              schema:
                $ref: '#/definitions/Thing'
            `,
			&BodyParameter{
				ParameterFields: ParameterFields{In: "body", Name: "body"},
				Schema:          &ReferenceSchema{Reference: "#/definitions/Thing"},
			},
		},

		{
			"string",
			`
            - name: param
              in: query
              type: string
            `,
			&StringParameter{ParameterFields: ParameterFields{In: "query", Name: "param"}},
		},
		{
			"number",
			`
            - name: param
              in: query
              type: number
            `,
			&NumberParameter{ParameterFields: ParameterFields{In: "query", Name: "param"}},
		},
		{
			"integer",
			`
            - name: param
              in: query
              type: integer
            `,
			&IntegerParameter{ParameterFields: ParameterFields{In: "query", Name: "param"}},
		},
		{
			"boolean",
			`
            - name: param
              in: query
              type: boolean
            `,
			&BooleanParameter{ParameterFields: ParameterFields{In: "query", Name: "param"}},
		},
		{
			"array",
			`
            - name: param
              in: query
              type: array
              items:
                type: string
            `,
			&ArrayParameter{
				ParameterFields: ParameterFields{In: "query", Name: "param"},
				ArrayItem:       ArrayItem{Items: &StringItem{}},
			},
		},
		{
			"$ref",
			`
			- $ref: '#/parameters/TestParam'
			`,
			&ReferenceParamter{Reference: "#/parameters/TestParam"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var out Parameters
			if err := yaml.Unmarshal([]byte(dedent.Dedent(tc.in)), &out); err != nil {
				t.Fatal("could not unmarshal. got error:", err)
			}
			if !reflect.DeepEqual(out[0], tc.out) {
				t.Error("Wrong value unmarshaled. got:", out[0], "expected:", tc.out)
			}
		})
	}
}

func TestUnmarshalArrayItem(t *testing.T) {
	tcs := []struct {
		name string
		in   string
		out  Parameter
	}{
		{
			"string",
			`
            - name: param
              in: query
              type: array
              items:
                type: string
            `,
			&ArrayParameter{
				ParameterFields: ParameterFields{In: "query", Name: "param"},
				ArrayItem:       ArrayItem{Items: &StringItem{}},
			},
		},
		{
			"number",
			`
            - name: param
              in: query
              type: array
              items:
                type: number
            `,
			&ArrayParameter{
				ParameterFields: ParameterFields{In: "query", Name: "param"},
				ArrayItem:       ArrayItem{Items: &NumberItem{}},
			},
		},
		{
			"integer",
			`
            - name: param
              in: query
              type: array
              items:
                type: integer
            `,
			&ArrayParameter{
				ParameterFields: ParameterFields{In: "query", Name: "param"},
				ArrayItem:       ArrayItem{Items: &IntegerItem{}},
			},
		},
		{
			"boolean",
			`
            - name: param
              in: query
              type: array
              items:
                type: boolean
            `,
			&ArrayParameter{
				ParameterFields: ParameterFields{In: "query", Name: "param"},
				ArrayItem:       ArrayItem{Items: &BooleanItem{}},
			},
		},
		{
			"array",
			`
            - name: param
              in: query
              type: array
              items:
                type: array
                items:
                  type: boolean
            `,
			&ArrayParameter{
				ParameterFields: ParameterFields{In: "query", Name: "param"},
				ArrayItem: ArrayItem{Items: &ArrayItem{
					Items: &BooleanItem{},
				}},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var out Parameters
			if err := yaml.Unmarshal([]byte(dedent.Dedent(tc.in)), &out); err != nil {
				t.Fatal("could not unmarshal. got error:", err)
			}
			if !reflect.DeepEqual(out[0], tc.out) {
				t.Error("Wrong value unmarshaled. got:", out[0], "expected:", tc.out)
			}
		})
	}
}

func TestResponsesUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
    200:
      description: success
      schema:
        type: string
    204:
      description: success with no content
    default:
      schema:
        $ref: '#/definitions/Error'
	`)

	var out Responses
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	expected := Responses{
		Default: &Response{Schema: &ReferenceSchema{Reference: "#/definitions/Error"}},
		Codes: map[int]Response{
			200: {
				Description: "success",
				Schema:      &StringSchema{},
			},
			204: {
				Description: "success with no content",
			},
		},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Error("Wrong value unmarshaled. got:", out, "expected:", expected)
	}
}

func TestObjectSchemaUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
    type: object
    properties:
      name:
        type: string
      version:
        type: number
	`)

	var out ObjectSchema
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	expected := ObjectSchema{
		Properties: &SchemaMap{
			{Name: "name", Schema: &StringSchema{}},
			{Name: "version", Schema: &NumberSchema{}},
		},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Error("Wrong value unmarshaled. got:", out, "expected:", expected)
	}

}

func TestSchemaMapUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
    name:
      type: string
    version:
      type: number
    active:
      type: boolean
    count:
      type: integer
    props:
      type: object
      properties:
        name:
          type: string
	`)

	var out SchemaMap
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	expected := SchemaMap{
		{Name: "name", Schema: &StringSchema{}},
		{Name: "version", Schema: &NumberSchema{}},
		{Name: "active", Schema: &BooleanSchema{}},
		{Name: "count", Schema: &IntegerSchema{}},
		{Name: "props", Schema: &ObjectSchema{
			Properties: &SchemaMap{
				{Name: "name", Schema: &StringSchema{}},
			},
		}},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Error("Wrong value unmarshaled. got:", out, "expected:", expected)
	}
}

func TestAllOfSchemaUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
    val:
      allOf:
        - type: object
          properties:
            name:
              type: string
        - type: object
          properties:
            age:
              type: integer
	`)

	var out SchemaMap
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	expected := &AllOfSchema{
		AllOf: []Schema{
			&ObjectSchema{Properties: &SchemaMap{
				{Name: "name", Schema: &StringSchema{}},
			}},
			&ObjectSchema{Properties: &SchemaMap{
				{Name: "age", Schema: &IntegerSchema{}},
			}},
		},
	}
	if !reflect.DeepEqual(out[0].Schema, expected) {
		t.Error("Wrong value unmarshaled. got:", out[0].Schema, "expected:", expected)
	}
}

func TestArraySchemaUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
    val:
      type: array
      items:
        type: string
	`)

	var out SchemaMap
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	expected := &ArraySchema{
		Items: &StringSchema{},
	}
	if !reflect.DeepEqual(out[0].Schema, expected) {
		t.Error("Wrong value unmarshaled. got:", out[0].Schema, "expected:", expected)
	}
}

func TestSecuritySchemeMapUnmarshalYAML(t *testing.T) {
	d := dedent.Dedent(`
    basicType:
      type: basic
      description: basic auth
    apiKeyType:
      type: apiKey
      description: api key auth
      name: Authorization
      in: header
    oauthType:
      type: oauth2
      description: oauth auth
      flow: implicit
      authorizationUrl: "http://whereever.place"
      scopes: {}
	`)

	var out SecuritySchemeMap
	if err := yaml.Unmarshal([]byte(d), &out); err != nil {
		t.Fatal("could not unmarshal. got error:", err)
	}

	bdesc := "basic auth"
	adesc := "api key auth"
	odesc := "oauth auth"
	oURL, _ := url.Parse("http://whereever.place")
	expected := SecuritySchemeMap{
		"basicType": &BasicSecurityScheme{SecuritySchemeFields{Description: &bdesc}},
		"apiKeyType": &APIKeySecurityScheme{
			SecuritySchemeFields: SecuritySchemeFields{Description: &adesc},
			Name:                 "Authorization",
			In:                   "header",
		},
		"oauthType": &OAuth2SecurityScheme{
			SecuritySchemeFields: SecuritySchemeFields{Description: &odesc},
			Flow:                 "implicit",
			AuthorizationURL:     oURL,
			Scopes:               make(map[string]string),
		},
	}
	if !reflect.DeepEqual(out, expected) {
		t.Error("Wrong value unmarshaled. got:", out, "expected:", expected)
	}
}
