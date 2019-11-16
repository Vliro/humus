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
func makeSchema(output io.Writer) {
	var buffer bytes.Buffer
	for name, val := range globalFields {
		//Emit types first.
		var ib bytes.Buffer
		fl: for _,v := range val.AllFields {
			for _,val := range v.Directives {
				if val.Name.Name == "hasInverse" {
					if v.flags & flagArray != 0 {
						continue fl

					}
				}
			}
			ib.WriteString(fmt.Sprintf(typeFieldDecl, v.Tag, v.getDgraphSchema(true)))
		}
		buffer.WriteString(fmt.Sprintf(typeDecl, name, ib.String()))
		ib = bytes.Buffer{}
	}
	for _, val := range globalFields {
		//Emit types first.
		var written = make(map[string]struct{})
		var ib bytes.Buffer
		fieldLoop: for _,v := range val.AllFields {
			if _,ok := written[v.Name]; ok {
				continue
			}
			var directives bytes.Buffer
			for _,dir := range v.Directives {
				switch dir.Name.Name {
				case "hasInverse":
					if v.flags & flagArray == 0 {
						directives.WriteString("@reverse ")
					} else {
						break fieldLoop
					}
				case "search":
					directives.WriteString("@index(")
					for _,v := range dir.Args {
						str := v.Value.String()
						directives.WriteString(str[1:len(str)-1])
					}
					directives.WriteString(") ")
				}
			}
			written[v.Name] = struct{}{}
			ib.WriteString(fmt.Sprintf(schemaDecl, "<"+v.Tag+">", v.getDgraphSchema(false), directives.String()))
		}
		_, _ = io.Copy(&buffer, &ib)
		ib = bytes.Buffer{}
	}
	io.Copy(output, &buffer)
}

//returns name + type
func (f *Field) getDgraphSchema(forType bool) string {
	var typ string
	switch {
	case f.flags & flagArray > 0:
		typ = "[" + toDgraphType(f.Type, f.flags, forType) + "]"
	default:
		typ = toDgraphType(f.Type, f.flags, forType)
	}
	return typ
}

func toDgraphType(str string, flag flags, forType bool) string {
	if flag & flagObject != 0 && !forType {
		return "uid"
	}
	if flag & flagEnum != 0 && !forType{
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