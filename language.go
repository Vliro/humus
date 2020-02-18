package humus

//Language represents a language that sets relevant queries to the language as specified.
type Language string

//A list of possible languages.
//TODO: Do not make these static constants. Allow arbitrary languages.
const (
	LanguageEnglish = "en"
	LanguageGerman  = "de"
	LanguageSwedish = "se"
	//Same as english.
	LanguageNone = ""
)
