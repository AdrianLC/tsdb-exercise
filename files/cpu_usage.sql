-- CREATE DATABASE homework;
-- \c homework
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE TABLE IF NOT EXISTS cpu_usage(
  ts    TIMESTAMPTZ NOT NULL,
  host  TEXT NOT NULL,
  usage DOUBLE PRECISION NOT NULL
);
SELECT create_hypertable('cpu_usage', 'ts');
