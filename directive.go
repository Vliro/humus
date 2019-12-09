package humus

//Directive is simply all possible Dgraph directives.
type Directive string

const (
	Cascade      Directive = "cascade"
	Normalize    Directive = "normalize"
	IgnoreReflex Directive = "ignorereflex"
)
