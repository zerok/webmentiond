package frontend

import "embed"

//go:embed dist css index.html demo.html
var FS embed.FS
