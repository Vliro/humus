package mulbase

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

//A list of Fields associated with an object.
type FieldList []Field

//TODO: offset
type CountType string

const (
	CountFirst  CountType = "first"
	CountOffset CountType = "offset"
	CountAfter  CountType = "after"
)

type AggregateType string

const (
	TypeCount AggregateType = "count"
	TypeSum   AggregateType = "sum"
	TypeVar   AggregateType = "var"
	TypeVal   AggregateType = "val"
)

type Count struct {
	Type  CountType
	Value int
}

// Field is a recursive data struct which represents a GraphQL query field.
type Field struct {
	Name        string
	SchemaField *SchemaField
	Fields      []Field
	Facet       bool
}



// MakeField constructs a Field of given name and returns the Field.
func MakeField(name string) Field {
	//TODO: better facet support
	fac := strings.Contains(name, "|")
	var x = Field{Name: name, SchemaField: getSchemaField(name), Facet: fac}
	return x
}

type FieldHolder struct {
	Fields []Field
}

func (fh *FieldHolder) String() string {
	var sb strings.Builder
	for _, v := range fh.Fields {
		writeField(v, &sb)
	}
	sb.WriteString(" uid ")

	return sb.String()
}

func writeField(f Field, sb *strings.Builder) {
	sb.WriteString(f.Name)
	if len(f.Fields) != 0 {
		sb.WriteString("{")
		for _, v := range f.Fields {
			writeField(v, sb)
			sb.WriteByte(' ')
		}
		sb.WriteString(" uid ")
		sb.WriteString("} ")
	}
	sb.WriteString(" ")
}

func (fh *FieldHolder) Copy() *FieldHolder {
	f := FieldHolder{}
	f.Fields = make([]Field, len(fh.Fields))
	for k, v := range fh.Fields {
		c := v
		f.Fields[k] = c.CopyFields()
	}
	return &f
}

func (fh *FieldHolder) FieldByName(name string) *Field {
	for _, v := range fh.Fields {
		if len(v.Name) > 0 && v.Name == name {
			return &v
		}
	}
	return nil
}

func (fh *FieldHolder) Add(name string) *FieldHolder {
	for _, v := range fh.Fields {
		if v.Name == name {
			return fh
		}
	}
	f := Field{}
	f.Name = name
	fh.Fields = append(fh.Fields, f)
	return fh
}

func (fh *FieldHolder) AddFieldStrings(in *FieldHolder) *FieldHolder {
	for k := range in.Fields {
		ex := false
		for i := range fh.Fields {
			if fh.Fields[i].Name == in.Fields[k].Name {
				ex = true
			}
		}
		if !ex {
			fi := Field{}
			fi.Name = in.Fields[k].Name
			fh.Fields = append(fh.Fields, fi)
		}
	}
	return fh
}

func MakeFieldHolder(s []string) *FieldHolder {
	var f []Field
	f = make([]Field, len(s))
	for k, v := range s {
		ft := MakeField(v)
		f[k] = ft
	}
	return &FieldHolder{Fields: f}
}

func CreateSubHolder(sub *FieldHolder, name string) FieldHolder {
	var farr []Field
	f := Field{}
	f.Name = "uid"
	farr = append(farr, f)
	farr = append(farr, Field{Fields: sub.Fields, Name: name})
	return FieldHolder{Fields: farr}
}

func (f *FieldHolder) Append(fh *FieldHolder, name string) *FieldHolder {
	//should copy the fieldholder
	fi := MakeField(name)
	newf := f.Copy()
	for _, v := range fh.Fields {
		c := v
		fi.Fields = append(fi.Fields, c)
	}
	newf.Fields = append(newf.Fields, fi)
	return newf
}

func (f *Field) CopyFields() Field {
	ret := Field{}
	ret.SchemaField = f.SchemaField
	ret.Name = f.Name
	ret.Facet = f.Facet
	ret.Fields = make([]Field, len(f.Fields))
	for k := range f.Fields {
		ret.Fields[k] = f.Fields[k].CopyFields()
	}
	return ret
}

// Implement fieldContainer
func (f *Field) getFields() *[]Field {
	return &f.Fields
}

func (f *Field) setFields(fs []Field) {
	f.Fields = fs
}

// String returns read only string token channel or an error.
// It checks if there is a circle.
func (f *Field) String(q *GeneratedQuery, parent string, sb *bytes.Buffer) error {
	if err := f.check(q); err != nil {
		// return a closed channel instead of nil for receiving from nil blocks forever, hard to debug and confusing to users.
		return errors.WithStack(err)
	}
	f.string(q, parent, sb)
	return nil
}

func (f *Field) writeLanguageTag(sb *bytes.Buffer, l Language) {
	if l != LanguageDefault && l != LanguageNone {
		sb.WriteString("@"+string(l)+":.")
	}
}

func (f *Field) getName() string {
	if f.Name[0] == '~' {
		return f.Name[1:]
	}
	return f.Name
}
// One may have noticed that there is a public String and a private string.
// The different being the public method checks the validity of the Field structure
// while the private counterpart assumes the validity.
// Returns whether this field is a facet field.
func (f *Field) string(q *GeneratedQuery, parent string, sb *bytes.Buffer) bool {
	if f.SchemaField == nil {
		f.SchemaField = getSchemaField(f.getName()	)
		if f.SchemaField == nil {
			panic(errors.Errorf("no schema %s", f.Name))
		}
	}
	if f.Facet {
		return true
	}
	agg, ok := q.FieldAggregate[parent]
	if ok {
		sb.WriteString(agg.Alias)
		sb.WriteString(" : ")
		sb.WriteString(string(agg.Type))
		sb.WriteString(tokenLP)
		sb.WriteString(f.Name)
		if f.SchemaField.Lang {
			f.writeLanguageTag(sb, q.Language)
		}
		sb.WriteString(tokenRP)
	} else {
		sb.WriteString(f.Name)
		if f.SchemaField.Lang {
			f.writeLanguageTag(sb, q.Language)
		}
	}

	//Handle the (first: -1)
	val, ok := q.FieldCount[parent]
	if ok {
		sb.WriteString(tokenLP)
		for k, v := range val {
			if k != 0 {
				sb.WriteString(tokenComma)
			}
			sb.WriteString(string(v.Type))
			sb.WriteString(tokenColumn)
			sb.WriteString(strconv.Itoa(v.Value))
		}
		sb.WriteString(tokenRP)
	}
	order, ok := q.FieldOrderings[parent]
	if ok {
		sb.WriteString(tokenLP)
		for k, v := range order {
			sb.WriteString(v.String())
			if k != 0 {
				sb.WriteString(tokenComma)
			}
		}
		sb.WriteString(tokenRP)
	}
	filter, ok := q.FieldFilters[parent]
	if ok {
		sb.WriteString(tokenSpace)
		filter.stringChan(q, parent, sb)
	}
	var tmp = new(bytes.Buffer)
	writefct := false
	if len(f.Fields) > 0 {
		tmp.WriteString(tokenLB)
		for i, field := range f.Fields {
			if len(field.Name) > 0 {
				if i != 0 && i != len(f.Fields) {
					tmp.WriteString(tokenSpace)
				}
				field.string(q, parent+"/"+field.Name, tmp)
			}
			if field.Facet && !writefct {
				sb.WriteString(tokenSpace)
				sb.WriteString("@facets")
				sb.WriteString(tokenSpace)
				writefct = true
			}
		}
		tmp.WriteString(tokenSpace)
		tmp.WriteString("uid")
		tmp.WriteString(tokenSpace)
		tmp.WriteString(tokenRB)
	}
	io.Copy(sb, tmp)
	return false
}

func (f *Field) check(q *GeneratedQuery) error {
	return nil
}


func MakeFields(s ...string) []Field {
	var f []Field
	f = make([]Field, len(s))
	for k, v := range s {
		ft := MakeField(v)
		f[k] = ft
	}
	return f
}

func (f *Field) AddFieldsBasic(str []string) *Field {
	for _, v := range str {
		t := Field{}
		t.Name = v
		f.Fields = append(f.Fields, t)
	}
	return f
}

func (f *Field) AddField(fs Field) *Field {
	f.Fields = append(f.Fields, fs)
	return f
}

// SetFields sets the sub fields of a Field and return the pointer to this Field.
func (f *Field) SetFields(fs ...Field) *Field {
	f.Fields = fs
	return f
}
