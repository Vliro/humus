package parse

import "github.com/Vliro/mulbase/gen/graphql-go/common"

//Field represents the parsed graphQL field alongside all
//definitions needed for generation.
type Field struct {
	//For  DB name.
	Tag string
	//This include omitempty for instance. it is related to the tag.
	WrittenTag string
	//Name for the field.
	Name  string
	//What type is this field? String etc.
	Type  string
	//Metadata.
	flags flags
	//like []*string. Used in templates.
	TypeLabel     string
	//did this come from the field or from the interface?
	FromInterface bool
	Nosave bool
	Directives []*common.Directive
	Nofield bool
	//For templates.
	Parent string
	IsArray bool
}

func (f *Field) HasDirective(name string) *common.Directive {
	for _,v := range f.Directives {
		if v.Name.Name == name {
			return v
		}
	}
	return nil
}

func (f *Field) IsScalar() bool {
	for _,v := range builtins {
		if v == f.Type {
			return true
		}
	}
	return f.flags & flagEnum > 0
}