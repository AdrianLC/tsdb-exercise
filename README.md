# Benchmark Exercise

## Prerequisites

Docker and docker compose installed in your system.

To execute outside of docker: golang 1.22

## Set up

Create a `.env` file as a way to provide the database URL with credentials to docker compose:

```
DB_DSN=postgres://<USER>:<PASSWORD>@<HOSTNAME>.cloud.timescale.com:35688/tsdb?sslmode=disable
```

To create the schema and insert the test data:
```
docker compose run psql ./setup-data.sh
```

## Commands

To print help about command usage:
```
docker compose run benchmark -help
```

To read the CSV from standard input:
```
docker compose run -T benchmark < ./files/query_params.csv
```

To provide the CSV from a file path (the files are mounted at `/app/files` in the docker container):
```
docker compose run benchmark -csv ./files/query_params.csv
```

To specify the number of concurrent workers running queries (defauts to 10):
```
docker compose run benchmark -workers 10 -csv ./files/query_params.csv
```

To run the tests within the docker image:
```
docker compose run --workdir /app/src --entrypoint /usr/local/go/bin/go benchmark test .
```

## Implementation Notes

### Connection Pool

Using a connection pool from pgx has been a must because as documented in https://github.com/jackc/pgx/wiki/Getting-started-with-pgx

> The *pgx.Conn returned by pgx.Connect() represents a single connection and is not concurrency safe. This is entirely appropriate for a simple command line example such as above. However, for many uses, such as a web application server, concurrency is required. To use a connection pool replace the import github.com/jackc/pgx/v5 with github.com/jackc/pgx/v5/pgxpool and connect with pgxpool.New() instead of pgx.Connect().

Alternatively we could have a separate connection for each worker but this would mean the program can only run as many workers as available connections.

### max_connections

I increased the `max_connections` parameter in my testing instance from 25 to 100. The number of connections available by default seems to be only 6 which is not that much and I wanted to test with more workers.

I found out about this with:

```
 tsdb=> show max_connections;
 max_connections 
 -----------------
 25
 (1 row)

 tsdb=> show superuser_reserved_connections ;
 superuser_reserved_connections 
 --------------------------------
 12
 (1 row)

 tsdb=> select application_name, count(*) from pg_stat_activity where state ='active' or state ='idle' group by 1;
             application_name             | count 
 -----------------------------------------+-------
  ForgeExplorer                           |     2
  Patroni heartbeat                       |     1
  Patroni restapi                         |     1
  TimescaleDB Background Worker Scheduler |     1
  postgres_exporter                       |     3
  psql                                    |     1
 (6 rows)
```

The pool only managed to acquire 5 or 6 connections. 25 - 12 - 9 (sum of above) puts as close to that number.
My conclusion  is there are maintenance tasks running too frequently to have enough connections with the limit of 25.

### Routing hostnames to the same worker

I implemented this with a hash of the hostname and modulues of the number of workers. A similar solution would be to parse the number suffix on the hostname values and do modulus with that instead. However it's bad to assume all hostnames would have numbers and it's not mentioned in the requirements. With a hash the solution works for any value that can be converted to bytes. Adding another column for example Account Name and routing by it instead would be a one line change.

### Statistics

#### Limitations

There is a memory limitation in how many parameters the program can handle. Although it's more than enough for the 200 in `query_params.csv`.
If we wanted to scale to thousands or millions of rows in the CSV there would be too much memory usage. This is because having an exact **median** statistic requires an ordered list of all the time values for every query. In a real world scenario usually buckets of values are used instead and observability apps rely on estimates for quantiles. Some details about this in https://dyladan.me/histograms/2023/05/03/histograms-vs-summaries/

#### Futher Improvements

Some change we could also monitor that I think are useful:

**rows returned for each query**

With this we might find a correlation with query times. The max query time might consistently match a hostname and time range with more data, so it's expected that scanning it to Go structs will take longer.

**percentiles 95, 90, 75**

Currently we have the median (p50), the average and the maximum query times. But this doesn't tell us how many queries have been much slower than the others. It could be that we have very few queries that takes 120ms or perhaps many. If we had several high percentiles we could compare them to know this.
For example if the max query time is 120ms, p95 100ms, p75 80ms and p50 50ms we know there are many queries taking over 80ms and 100ms.
However if we had max query time 120ms, p95 70ms and p50 50ms we know that query that takes 120ms is an outlier while most of them are faster than 70ms.

**separate time measures for server query time and client fetching and scanning**

This is about measuring the time to read rows and scanning them separately from the query time. With this you could see how much time is from drivers or client code versus the server resolving results.
