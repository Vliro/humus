## The repository

Mulbase/Mulgen is a front-end library for DGraph. It includes static (mostly) code generation based on the same schema supplied to the database so no need to keep the client structs separate. 

Note that this code does not, as of right, have any way to go get in a regular manner. It exists simply in go/src but hopefully this will be changed in due time. There are no guarantees whatsoever about this library right now.

## Mulgen

Mulgen is the library for generating structs and common definitions.

## Mulbase

Mulbase represents the regular code for interacting with the database through the DNode interface which is satisfied by all generated classes. It includes GeneratedQuery's, StaticQuery's(simple string queries), SingleMutations as well as MultipleMutations. These all exist inside Txn objects with simple commit-now calls also available straight in the *DB class.

Right now, it does not have proper testing. What needs to be added is a method to autostart the GraphQL DGraph api, alter the schema,

It supports(TODO) DGraph language tags by following the schema and supplying them in the generated query as needed.

An underlying point about many functions like SaveScalars is it does not run immediately. It returns a query object as you might want to extend the query. After that you simply
execute it inside a transaction object.

## Notes about GraphQL

The GraphQL Non-null(! operator) defines whether or not whether scalar values are autogenerated. For non-scalar values it doesn't really matter right now. 

## Graphql-Go

This is derived from https://github.com/graph-gophers/graphql-go. A lot of code has been removed that
was not necessary since primarily the parsing was necessary and not the actual GraphQL elements. 
The foremost change was making the schema and its objects public for code generation.

## Getting started

Running mulbase/gen/parse/main.go with input/output flags generates the model files, models.go and gen.go. 
These are used with mulbase. Create a new DB object using mulbase.Init() and set the schema from the global fields
generated. 

## TODOS

Language tags are soon to be added properly, I'm not sure how they work with DGraph GraphQL. 
Proper testing. 
Finalize the API.
Comment.
Clean up the gen directory. Code is messy.