## Notes

This library is a work in progress and is in no way advised for production. 
Right now geometry is not available.

## The repository

Mulbase/gen is a front-end library for DGraph. It includes static (mostly) code generation based on the same schema supplied to the database.
It ensures json tags are set in sync with the schema and ensures compatibility with the database.

The schema is built using regular GraphQL syntax and generates a regular DGraph schema. 

Note that this code does not, as of right now, use vendor/modules. This will be changed eventually.

It allows for different methods of communicating with the database. This includes static queries, dynamic generated queries, 
interface mutations, custom mutations among others.

In the repository almost everything is interfaced. All generated structs adhere to the same interface to ensure you are aware
of what you are sending. 

## gen

gen is the library for generating structs and common definitions.

The code is mostly hacked together to get the functionality. Once I've decided exactly how it will generate models the code will
be written to be easier to understand and deal with. 

It supports using both regular DGraph schemas as well as its GraphQL api. It generates a schema that roughly matches (What I) believe 
is the appropriate dgraph schema and sets up the fields according to it. 

## Mulbase

Mulbase represents the regular code for interacting with the database through the DNode interface which is satisfied by all generated classes. It includes GeneratedQuery's, StaticQuery's(simple string queries), SingleMutations as well as MultipleMutations. These all exist inside Txn objects with simple commit-now calls also available straight in the *DB class.
Also, arbitrary interfaces can be mutated as well, for instance a map\[string\]interface{} using the Mapper wrapper over it as well as custom interfaces.
Most features are wrapped around an interface and by implementing methods you can modify how the package sends data to the DB. This level of control
is crucial for a graph database as, for instance, mutating a struct with an edge can cause accidental overwrite of relative values. By default, edges are never saved. This can
be overwritten through the Save interface as well as via manual mapping.

All queries and mutations returned are just a struct. They satisfy the Query and Mutate interface respectively and can be called into any Querier. Per default this includes DB and Txn, with
the difference being DB immediately commits all executions(or discards).

## Graphql-Go

This is derived from https://github.com/graph-gophers/graphql-go. A lot of code has been removed that
was not necessary since the parsing was needed and not the actual GraphQL elements. (It does not represent a GraphQL server)
The foremost change was making the schema and its objects public for code generation.

## Slices

References to other classes in the database is currently modeled as []Type rather than []*Type. This simplifies allocations but 
causes reference issues. For instance, pointer to &array at index k keeps the entire array alive. Therefore, it might be better
to use a slice of pointers. 

## Variables

All inputs into functions are transformed into GraphQL variables to deny possible injection of data.

## Mutations

A guarantee is that for all mutations, only fields that are set will be mutated. This does cause an issue with default time.Time. Due to the way omitempty and structs behave
it is now a pointer. 

## Getting started

Running mulbase/gen/parse/main.go with input/output flags generates the model files, gen.go and enum.go. 
An example instance can be seen under the examples folder. 

## TODO

Dgraph has a type system. Right now only top level nodes have their type set. In an instance where you save multiple levels in one mutation
the behaviour might be that it automatically traverses all edges and sets their type if their UID is non-empty. This is not decided.