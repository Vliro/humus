package parse

//Non-file specific 'constants'.

//list of template files.
var templates = map[string]string{
	"Recurse": "recurse.template",
	"Field":   "global.template",
	"Async":   "async.template",
	"Model":   "model.template",
	"Enum":    "enum.template",
	"All":     "allfiles.template",
}

//Scalar builtin types.
var builtins = map[string]string{
	"String":   "string",
	"Int":      "int",
	"Boolean":  "bool",
	"DateTime": "time.Time",
	"Float":    "float64",
	//Uid handled separately.
	"ID": "",
}

//flags represent object metadata.
type flags int

const (
	flagNotNull = 1 << iota
	flagArray
	flagScalar
	flagPointer
	flagInterface
	flagEnum
	flagReverse
	flagObject
	flagFacet
	flagTwoWay
	flagLang
)
