#!/bin/bash

# Check if a migration name is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <migration_name>"
    exit 1
fi

# Get the current Unix timestamp
timestamp=$(date +%s)

# Create a migration file with the provided name and timestamp prefix
migration_name="$timestamp"_"$1.sql"

migration_file_name="migrations/$migration_name"

touch $migration_file_name

echo "Migration file created: $migration_name"
