package templates

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// p should not be used outside this file, instantiated once to avoid a lot of instantiations.
// defaulting now to english but the language should change depending on http Accept-Language headers or cookie
// configuration. Not aiming to support internationalization for now.
var p = message.NewPrinter(language.English)

// humanInt returns v as a formatted number in the english language
func humanInt(v int) string {
	return p.Sprint(v)
}
