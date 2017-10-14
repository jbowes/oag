package config

import (
	"io"
	"text/template"
)

var defaultConfig = template.Must(template.New("cfg").Parse(
	`# Default oag configuration file. Edit these values as appropriate
# Required: The openapi document to generate a client for.
document: {{.Document}}
package:
  # Required: The import path of your package.
  path: {{.Path}}
  # Optional: define a package name if it is different from the import path
  # name: {{.Name}}

# Optional mapping of definitions to types.
# types:
#   SomeDefinedType: github.com/org/package.TypeName

# Optional mapping of formats for strings to types. Types must implement
# the encoding.TextMarshaler/encoding.TextUnmarshaler interfaces.
# string_formats:
#   telephone: github.com/org/package.TelephoneNumber
`))

// WriteDefaultConfig writes a default configuration to the given io.Writer.
func WriteDefaultConfig(f io.Writer) error {
	return defaultConfig.Execute(f, map[string]string{
		"Document": "openapi.yaml",
		"Path":     "github.com/yourorg/yourpackage",
		"Name":     "yourpackage",
	})
}
