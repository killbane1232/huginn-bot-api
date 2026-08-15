#!/bin/sh
set -eu

if [ "$(id -u)" = "0" ]; then
    install -d -o bot -g bot /app/data /app/data/uploads
    chown -R bot:bot /app/data
    exec gosu bot "$@"
fi

exec "$@"
