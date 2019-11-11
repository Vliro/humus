package parse
//TODO: Natively integrate this into the parser but for now this works.
const header = `directive @search(
  by: [String]
) on FIELD_DEFINITION | ENUM_VALUE

directive @hasInverse(
  field: String
) on FIELD_DEFINITION | ENUM_VALUE
`
