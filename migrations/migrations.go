package migrations

import "embed"

// FS embeds all SQL migration files for goose auto-migration.
//go:embed postgres/*.sql sqlite/*.sql
var FS embed.FS
