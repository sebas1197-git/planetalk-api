// Package migrations embeds the SQL migration files so they can be applied from
// anywhere (e.g. integration tests) without depending on the working directory.
// Production startup still reads them from disk via internal/db.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
