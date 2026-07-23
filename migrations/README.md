# Database Migrations

Place SQL migration files here with the following naming convention:

- `001.up.sql` — Up migration for version 1
- `001.down.sql` — Down migration for version 1

Migrations are applied in version order. Embedded migrations (in `internal/database/migrations.go`) take priority; file-based migrations supplement them.
