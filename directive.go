package humus

//Directive contains all possible query directives for dgraph.
//These are applied at the root of the query.
type Directive string

const (
	Cascade      Directive = "cascade"
	Normalize    Directive = "normalize"
	IgnoreReflex Directive = "ignorereflex"
)
