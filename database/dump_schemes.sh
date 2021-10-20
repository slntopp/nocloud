#!/bin/bash

arangodump --server.database nocloud --overwrite --server.password "$ARANGO_ROOT_PASSWORD" --include-system-collections true

