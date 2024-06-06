# Instructions

## Prerequisites

Docker and docker compose installed in your system.

## Set up

Create a `.env` file as a way to provide the database URL with credentials to docker compose:

```
DB_DSN=postgres://<USER>:<PASSWORD>@<HOSTNAME>.cloud.timescale.com:35688/tsdb?sslmode=disable
```

To create the schema and write the test data:
```
docker compose run psql setup-data.sh
```

## Commands

To read the CSV from standard input:
```
docker compose run -T benchmark < ./files/query_params.csv
```

To provide the CSV as file path (the files are mounted at `/app/files` in the docker container):
```
docker compose run benchmark -csv ./files/query_params.csv
```

To run tests within the docker image:
```
docker compose run --workdir /app/src --entrypoint /usr/local/go/bin/go benchmark test .
```

