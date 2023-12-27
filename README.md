# Flywheel

Unity client project: https://github.com/cbodonnell/flywheel-client

## Development

Start the server with:
```
docker compose up
```

Pass the `--build` flag to rebuild the image.

To stop the server, run:
```
docker compose down
```

Pass the `-v` flag to remove volumes.

## Schema Migrations

To create a new migration, change to the `schema` directory and run:
```
./scripts/create_migration.sh <migration_name>
```

This will create a new migration file in the `migrations` directory. Edit this file to add the migration.

TODO: Switch to a more robust migration tool. Something like [goose](https://github.com/pressly/goose)?
