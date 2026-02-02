package migrations

import "embed"

// Note: Must match the default schema in Postgres unless overridden.
const Schema = "public"

// FS embeds all SQL migrations in this directory.
//
//go:embed *.sql
var FS embed.FS
