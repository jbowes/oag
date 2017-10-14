package pkg

// Visibility is an enum that describes a type, method, or field's configured
// Visibility in the generted code
type Visibility string

// The levels of visibility
const (
	Public   Visibility = "public"
	Private  Visibility = "private"
	Disabled Visibility = "disabled" // don't even generate this code
)
