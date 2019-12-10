
# Get started  
  Install the code generator using go get.
```  
go get github.com/Vliro/humus/gen  
```  
  
# Specify a schema  
  
```  
#spec/schema.graphql
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
# Generate code  
  Spec is an example folder where relevant GraphQL files lie and models is just an example output.
```shell script  
$GOPATH/bin/gen -input=spec/ -output=models/
```  
  ---
```  
#schema.txt
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
Code below is an extract from the generated code.  
```go  
//models/models.go
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
Run an example query.  
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
  
# Understand the code  
  
gen is the folder containing all code pertaining to generating Go code from GraphQL specifications.  
testing contains all test code.  
//WIP examples contains example usages of the library.  
  
All code is documented over at Godoc.  
https://godoc.org/github.com/Vliro/humus