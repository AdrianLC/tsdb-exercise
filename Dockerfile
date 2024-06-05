FROM golang:1.21.10

WORKDIR /app

COPY * .

RUN go build -o benchmark main.go && \
    chmod +x benchmark

ENTRYPOINT ["/app/benchmark"]
