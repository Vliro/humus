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
type Post struct {  
   //This line declares basic properties for a database node.  
  humus.Node  
  Text          string `json:"Post.text,omitempty"`  
  DatePublished *time.Time `json:"Post.datePublished,omitempty"`  
}
//...
```
# Run a query
```go
//Gets question top level values as well as the edge Question.from
var getFields = QuestionFields.Sub(QuestionFrom, UserFields)

func GetAllPosts() ([]*Post, error) {
	var p []*Post
	var q = humus.NewQuery(getFields).
		Function(humus.Type).Value("Post")
	err := db.Query(context.Background(), q, &p)
	return p, err
}
```

# Learn
Read the testing folder and the (work in progress) wiki for exact information.