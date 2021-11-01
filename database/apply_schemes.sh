#!/bin/bash

arangorestore --create-database true --import-data false --server.password "$ARANGO_ROOT_PASSWORD"