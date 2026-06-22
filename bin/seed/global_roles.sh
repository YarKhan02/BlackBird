#!/bin/bash
set -euo pipefail

# CONNECTION_URL="postgresql://yarkhan:yarkhanworkshop@localhost:5432/blackbird"
CONNECTION_URL="postgresql://neondb_owner:npg_5oygIUDJwB8d@ep-patient-resonance-aokqkwv1-pooler.c-2.ap-southeast-1.aws.neon.tech/neondb"

if [ -z "$CONNECTION_URL" ]; then
  echo "Usage: $0 <database_url> or set DATABASE_URL" >&2
  exit 1
fi

psql "$CONNECTION_URL" -v ON_ERROR_STOP=1 <<'SQL'
INSERT INTO global_roles (id, name) VALUES
  (1, 'super_admin'),
  (2, 'admin'),
  (3, 'user'),
  (4, 'banned')
ON CONFLICT DO NOTHING;
SQL