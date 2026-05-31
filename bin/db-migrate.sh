#!/bin/bash

# Colors
CYAN='\033[1;36m'
GREEN='\033[1;32m'
RED='\033[1;31m'
NO_COLOR='\033[0m'
LABEL="db-migrate"

printf "${CYAN}== ${LABEL}${NO_COLOR}\n"

# Set your connection string
CONNECTION_URL="postgres://yarkhan:yarkhanworkshop@localhost:5432/blackbird?sslmode=disable"

MIGRATIONS_PATH="./migrations"

# Run migrations
migrate -path "$MIGRATIONS_PATH" -database "$CONNECTION_URL" up

# Check status
if [ $? -eq 0 ]; then
  echo "Migrations applied successfully"
else
  echo "Migration failed"
  exit 1
fi