package mulbase

//All dgraph query directives.
type Directive string

const (
	Cascade      Directive = "cascade"
	Normalize    Directive = "normalize"
	IgnoreReflex Directive = "ignorereflex"
)
