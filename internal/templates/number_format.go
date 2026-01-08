package templates

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// p should not be used outside this file, instantiated once to avoid a lot of instantiations.
// defaulting now to english but the language should change depending on http Accept-Language headers or cookie
// configuration. Not aiming to support internationalization for now.
var p = message.NewPrinter(language.English)

// humanInt returns v as a formatted number in the english language.
func humanInt(v int) string {
	return p.Sprint(v)
}

func humanInt64(v int64) string {
	return p.Sprint(v)
}

// humanBytes returns v in a human-readable byte format with auto-scaling (B, KB, MB, GB, TB).
// Uses 1024 as the divisor for binary units.
// Examples: "2.3 MB", "524 KB", "1.2 GB"
func humanBytes(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
		tb = gb * 1024
	)

	absBytes := bytes
	if absBytes < 0 {
		absBytes = -absBytes
	}

	switch {
	case absBytes < kb:
		return fmt.Sprintf("%d B", bytes)
	case absBytes < mb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/kb)
	case absBytes < gb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/mb)
	case absBytes < tb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/gb)
	default:
		return fmt.Sprintf("%.1f TB", float64(bytes)/tb)
	}
}
