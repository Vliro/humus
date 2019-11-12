package parse

//Non-file specific 'constants'.

//list of template files.
var templates = map[string]string {
	"Get": "get.template",
	"Field": "global.template",
	"Async": "async.template",
	"Model": "model.template",
}

//Scalar builtin types.
var builtins = map[string]string{
	"String": "string",
	"Int":    "int",
	"Boolean": "bool",
	"DateTime": "time.Time",
	"Float": "float64",
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
)