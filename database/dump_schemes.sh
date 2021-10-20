#!/bin/bash

arangodump --server.database nocloud --overwrite --dump-data false --server.password "$ARANGO_ROOT_PASSWORD" --include-system-collections true

