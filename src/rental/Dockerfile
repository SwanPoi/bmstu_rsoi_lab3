FROM golang:1.25.1

WORKDIR /rental

COPY ./ ./

RUN go mod tidy
RUN go mod vendor
RUN go build -o main ./cmd/main.go

EXPOSE 8060
ENTRYPOINT [ "./main" ]
