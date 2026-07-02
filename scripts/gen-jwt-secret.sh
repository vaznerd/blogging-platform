#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-.env}"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "Error: $ENV_FILE does not exist."
  exit 1
fi

NEW_SECRET="$(openssl rand -base64 32)"

if grep -q '^JWT_SECRET=' "$ENV_FILE"; then
  sed -i "s|^JWT_SECRET=.*|JWT_SECRET=${NEW_SECRET}|" "$ENV_FILE"
  echo "Updated JWT_SECRET in $ENV_FILE"
else
  printf '\nJWT_SECRET=%s\n' "$NEW_SECRET" >>"$ENV_FILE"
  echo "Added JWT_SECRET to $ENV_FILE"
fi
