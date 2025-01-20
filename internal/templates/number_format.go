package templates

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var p = message.NewPrinter(language.English)

func humanInt(v int) string {
	return p.Sprint(v)
}
