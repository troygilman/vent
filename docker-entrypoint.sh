#!/bin/sh
set -eu

mkdir -p tmp

atlas migrate apply \
    --dir "file://examples/basic/ent/migrate/migrations" \
    --url "sqlite://tmp/test.db?_fk=1"

exec vent-example
