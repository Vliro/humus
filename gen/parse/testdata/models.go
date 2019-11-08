package mulgen

//Code generated by Mulbase. Do not edit but feel free to read comments.
import (
	"mulbase"
)

type Todo struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	Text string `json:"Todo.text"`
	Done bool   `json:"Todo.done"`
	User *User  `json:"Todo.user"`
}

var TodoFields mulbase.FieldList = []mulbase.Field{MakeField("Todo.text"), MakeField("Todo.done"), MakeField("Todo.user")}

type User struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	Name string `json:"User.name"`
}

var UserFields mulbase.FieldList = []mulbase.Field{MakeField("User.name")}

type Character struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	Name      string     `json:"Character.name"`
	AppearsIn []*Episode `json:"Character.appearsIn"`
}

var CharacterFields mulbase.FieldList = []mulbase.Field{MakeField("Character.name"), MakeField("Character.appearsIn")}

type Episode struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	Name string `json:"Episode.name"`
}

var EpisodeFields mulbase.FieldList = []mulbase.Field{MakeField("Episode.name")}

type Query struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	Todos []*Todo `json:"Query.todos"`
}

var QueryFields mulbase.FieldList = []mulbase.Field{MakeField("Query.todos")}
