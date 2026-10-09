#!/usr/bin/env bash
# Starts a throwaway local Postgres for tests and prints its URL. Usage: testDb.sh start|stop
set -euo pipefail
export LC_ALL=C
cd "$(dirname "$0")/.."
dataDir="$PWD/.tmp/pgdata"
port="${KEROKU_TEST_PG_PORT:-54329}"
case "${1:-start}" in
  start)
    if [ ! -f "$dataDir/PG_VERSION" ]; then
      mkdir -p "$dataDir"
      initdb -D "$dataDir" -U keroku --auth=trust >/dev/null
    fi
    if ! pg_ctl -D "$dataDir" status >/dev/null 2>&1; then
      pg_ctl -D "$dataDir" -l "$PWD/.tmp/pg.log" -o "-p $port -k $PWD/.tmp -c listen_addresses=127.0.0.1 -c max_connections=400" -w start >/dev/null
    fi
    echo "postgres://keroku@127.0.0.1:$port/postgres?sslmode=disable"
    ;;
  stop)
    pg_ctl -D "$dataDir" -m fast stop >/dev/null 2>&1 || true
    ;;
  *)
    echo "usage: $0 start|stop" >&2
    exit 2
    ;;
esac
