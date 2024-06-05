#!/bin/sh

psql ${DB_DSN} < cpu_usage.sql
psql ${DB_DSN} -c "\COPY cpu_usage FROM cpu_usage.csv CSV HEADER"
