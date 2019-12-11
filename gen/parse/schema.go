package parse

import (
	"bytes"
	"fmt"
	"io"
)

type objectType int

const (
	objectRegular objectType = iota
	objectInterface
)

type Object struct {
	Fields    []Field
	AllFields []Field
	Type      objectType
}

var globalFields = make(map[string]Object)

const typeDecl = `type %s {
%s}
`

const typeFieldDecl = "%s : %s  \n"

const schemaDecl = "%s: %s %s . \n"

//makeSchema generates a regular graphQL schema for dgraph.
//it is very much temporary!
func makeSchema(output io.Writer, g *Generator) {
	var buffer bytes.Buffer
	for name, val := range globalFields {
		//Emit types first.
		var ib bytes.Buffer
	fl:
		//Change this to globalFields to put all values in type declaration.
		for _, v := range val.Fields {
			if val := v.HasDirective("facet"); val != nil {
				continue fl
			}
			for _, val := range v.Directives {
				if val.Name.Name == "hasInverse" {
					if v.flags&flagArray != 0 {
						continue fl

					}
				}
			}
			ib.WriteString(fmt.Sprintf(typeFieldDecl, v.Tag, v.getDgraphSchema(true)))
		}
		buffer.WriteString(fmt.Sprintf(typeDecl, name, ib.String()))
		ib = bytes.Buffer{}
	}
	var written = make(map[string]struct{})
	for _, val := range globalFields {
		//Emit types first.
		var ib bytes.Buffer
	fieldLoop:
		for _, v := range val.AllFields {
			if _, ok := written[v.Tag]; ok {
				continue fieldLoop
			}
			var directives bytes.Buffer
			for _, dir := range v.Directives {
				switch dir.Name.Name {
				case "hasInverse":
					if v.flags&flagArray == 0 {
						directives.WriteString("@reverse ")
					} else {
						if val := v.HasDirective("source"); val != nil {
							directives.WriteString("@reverse ")
						} else {
							continue fieldLoop
						}
					}
				case "count":
					directives.WriteString("@count")
				case "search":
					directives.WriteString("@index(")
					for _, v := range dir.Args {
						str := v.Value.String()
						directives.WriteString(str[1 : len(str)-1])
					}
					directives.WriteString(") ")
				case "facet":
					continue fieldLoop
				case "lang":
					directives.WriteString("@lang ")
				}
			}
			written[v.Tag] = struct{}{}
			ib.WriteString(fmt.Sprintf(schemaDecl, "<"+v.Tag+">", v.getDgraphSchema(false), directives.String()))
		}
		_, _ = io.Copy(&buffer, &ib)
		ib = bytes.Buffer{}
	}
	_, err := io.Copy(output, &buffer)
	if err != nil {
		panic(err)
	}
}

//returns name + type
func (f *Field) getDgraphSchema(forType bool) string {
	var typ string
	switch {
	case f.flags&flagArray > 0:
		typ = "[" + toDgraphType(f.Type, f.flags, forType) + "]"
	//Enums are int type in the database. This is a special case for now.
	case f.flags&flagEnum > 0:
		typ = "int"
	default:
		typ = toDgraphType(f.Type, f.flags, forType)
	}
	return typ
}

func toDgraphType(str string, flag flags, forType bool) string {
	if flag&flagObject != 0 && !forType {
		return "uid"
	}
	//Artifact from GraphQL. Not sure yet.
	if flag&flagEnum != 0 && !forType {
		return "int"
	}
	switch str {
	case "int", "string":
		return str
	case "time.Time":
		return "datetime"
	case "float64":
		return "float"
	}
	return str
}
