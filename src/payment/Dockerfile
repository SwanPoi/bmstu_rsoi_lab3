FROM golang:1.25.1

WORKDIR /payment

COPY ./ ./

RUN go mod tidy
RUN go mod vendor
RUN go build -o main ./cmd/main.go

EXPOSE 8050
ENTRYPOINT [ "./main" ]
