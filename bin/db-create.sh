#!/bin/bash

CYAN='\033[1;36m'
NO_COLOR='\033[0m'
LABEL="db-create"
printf "${CYAN}== ${LABEL}${NO_COLOR}\n"

# Set your connection string
CONNECTION_URL="postgresql://yarkhan:yarkhanworkshop@localhost:5432/postgres"

# Create the new database
psql $CONNECTION_URL -c "CREATE DATABASE blackbird"