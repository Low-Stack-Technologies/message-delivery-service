package migrations

import "embed"

// FS embeds the SQL migrations so the binary can initialize the database without
// relying on the source tree being present at runtime.
//
//go:embed *.sql
var FS embed.FS
