# Specify a schema

```
interface Post {  
    text: String! @search (by:[hash])  
    datePublished: DateTime!  
}  
  
type Question implements Post {  
    title: String! @search (by:[hash])  
    from: User!  
}  
  
type User {  
    name: String!  
}
```
# Generate a schema
```
type Post {  
    Post.text : string    
    Post.datePublished : datetime    
}  
type Question {  
    Question.title : string    
    Question.from : User    
}  
type User {  
    User.name : string    
}  
<Post.text>: string @index(hash)  .   
<Post.datePublished>: datetime  .   
<Question.title>: string @index(hash)  .   
<Question.from>: uid  .   
<User.name>: string  . 
```
# Generate structs
```go
//...
type Question struct {
	//This line declares basic properties for a database node.
	humus.Node
	//List of interfaces implemented.
	Post
	//Regular fields
	Title string `json:"Question.title,omitempty"`
	From  *User  `json:"Question.from,omitempty"`
}
//...
```
# Run a query
```go
//Gets question top level values as well as the edge Question.from.
var getFields = QuestionFields.Sub(QuestionFrom, UserFields)
//Get all questions in the database.
func GetAllQuestions() ([]*Question, error) {
	var p []*Question
	var q = humus.NewQuery(getFields).
		Function(humus.Type).Value("Question")
	err := db.Query(context.Background(), q, &p)
	return p, err
}
```

# Learn
Read the testing folder and the (work in progress) wiki for exact information.