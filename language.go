package mulbase

type Language string

//A list of possible languages.
//TODO: Do not make these static constants. Allow arbitrary languages.
const (
	LanguageDefault = "en"
	LanguageSwedish = "se"
	//Same as english.
	LanguageNone = ""
)
