FROM golang:1.22.4

WORKDIR /app

RUN mkdir bin

WORKDIR /app/src 

COPY ./src/go.mod .
COPY ./src/go.sum .
RUN go mod download

COPY ./src/ .
RUN go build -o ../bin/benchmark . && \
    chmod +x ../bin/benchmark

WORKDIR /app
ENTRYPOINT ["/app/bin/benchmark"]
