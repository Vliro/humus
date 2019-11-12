package mulgen

// Code generated by mulgen. DO NOT EDIT (or feel free but it will be lost!).
import (
	"context"
	"mulbase"
	"time"
)

//Created from a GraphQL interface.
type Post struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	Text          string    `json:"Post.text"`
	DatePublished time.Time `json:"Post.datePublished"`
}

var PostFields mulbase.FieldList = []mulbase.Field{MakeField("Post.text", 0), MakeField("Post.datePublished", 0)}

//SaveValues saves the node values that
//do not contain any references to other objects.
func (r *Post) SaveValues(ctx context.Context, txn *mulbase.Txn) error {
	mut := mulbase.CreateMutation(r.Values(), mulbase.QuerySet)
	_, err := txn.RunQuery(ctx, mut)
	return err
}

//Fields returns all Scalar fields for this value.
func (r *Post) Fields() mulbase.FieldList {
	return PostFields
}

//Sets the types. This includes interfaces.
func (r *Post) SetType() {
	r.Type = []string{
		"Post",
	}
}

//Values returns all the scalar values for this node.
func (r *Post) Values() mulbase.DNode {
	var m PostScalars
	m.Text = r.Text
	m.DatePublished = r.DatePublished
	m.Uid = r.Uid
	return &m
}

//PostScalars is simply to avoid a map[string]interface{}
//It is a mirror of the previous struct with all scalar values.
type PostScalars struct {
	mulbase.Node
	Text          string    `json:"Post.text"`
	DatePublished time.Time `json:"Post.datePublished"`
}

func (s *PostScalars) Values() mulbase.DNode {
	return s
}

func (s *PostScalars) Fields() mulbase.FieldList {
	return PostFields
}

//End of model.template
type Question struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	//List of interfaces implemented.
	Post
	//Regular fields
	Title string `json:"Question.title"`
}

var QuestionFields mulbase.FieldList = []mulbase.Field{MakeField("Question.title", 0)}

//SaveValues saves the node values that
//do not contain any references to other objects.
func (r *Question) SaveValues(ctx context.Context, txn *mulbase.Txn) error {
	mut := mulbase.CreateMutation(r.Values(), mulbase.QuerySet)
	_, err := txn.RunQuery(ctx, mut)
	return err
}

//Fields returns all Scalar fields for this value.
func (r *Question) Fields() mulbase.FieldList {
	return QuestionFields
}

//Sets the types. This includes interfaces.
func (r *Question) SetType() {
	r.Type = []string{
		"Question",
		"Post",
	}
}

//Values returns all the scalar values for this node.
func (r *Question) Values() mulbase.DNode {
	var m QuestionScalars
	m.Title = r.Title
	m.Text = r.Text
	m.DatePublished = r.DatePublished
	m.Uid = r.Uid
	return &m
}

//QuestionScalars is simply to avoid a map[string]interface{}
//It is a mirror of the previous struct with all scalar values.
type QuestionScalars struct {
	mulbase.Node
	Title         string    `json:"Question.title"`
	Text          string    `json:"Post.text"`
	DatePublished time.Time `json:"Post.datePublished"`
}

func (s *QuestionScalars) Values() mulbase.DNode {
	return s
}

func (s *QuestionScalars) Fields() mulbase.FieldList {
	return QuestionFields
}

//End of model.template
type Comment struct {
	//This line declares basic properties for a database node.
	mulbase.Node
	//List of interfaces implemented.
	Post
	//Regular fields
	CommentsOn Post `json:"Comment.commentsOn"`
}

var CommentFields mulbase.FieldList = []mulbase.Field{MakeField("Comment.commentsOn", 0|mulbase.MetaObject)}

//SaveValues saves the node values that
//do not contain any references to other objects.
func (r *Comment) SaveValues(ctx context.Context, txn *mulbase.Txn) error {
	mut := mulbase.CreateMutation(r.Values(), mulbase.QuerySet)
	_, err := txn.RunQuery(ctx, mut)
	return err
}

//Fields returns all Scalar fields for this value.
func (r *Comment) Fields() mulbase.FieldList {
	return CommentFields
}

//Sets the types. This includes interfaces.
func (r *Comment) SetType() {
	r.Type = []string{
		"Comment",
		"Post",
	}
}

//Values returns all the scalar values for this node.
func (r *Comment) Values() mulbase.DNode {
	var m CommentScalars
	m.Text = r.Text
	m.DatePublished = r.DatePublished
	m.Uid = r.Uid
	return &m
}

//CommentScalars is simply to avoid a map[string]interface{}
//It is a mirror of the previous struct with all scalar values.
type CommentScalars struct {
	mulbase.Node
	Text          string    `json:"Post.text"`
	DatePublished time.Time `json:"Post.datePublished"`
}

func (s *CommentScalars) Values() mulbase.DNode {
	return s
}

func (s *CommentScalars) Fields() mulbase.FieldList {
	return CommentFields
}

//End of model.template
