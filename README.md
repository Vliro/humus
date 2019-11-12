## Mulbase

Mulbase/Mulgen is a front-end library for DGraph. It includes static (mostly) code generation based on the same schema supplied to the database so no need to keep the client structs separate. 

## Mulgen

Mulgen is the library for generating structs and common definitions.

## Mulbase

Mulbase represents the regular code for interacting with the database through the DNode interface which is satisfied by all generated classes. It includes GeneratedQuery's, StaticQuery's(simple string queries), SingleMutations as well as MultipleMutations. These all exist inside Txn objects with simple commit-now calls also available straight in the *DB class.

## Graphql-Go

This is derived from https://github.com/graph-gophers/graphql-go. A lot of code has been removed that
was not necessary since primarily the parsing was necessary and not the actual GraphQL elements. 
The foremost change was making the schema and its objects public for code generation.