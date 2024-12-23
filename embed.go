package namemyserver

import "embed"

//go:embed "db/migrations/*"
var MigrationsFS embed.FS

//go:embed "frontend/dist/*"
var FrontendFS embed.FS
