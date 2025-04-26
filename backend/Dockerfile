FROM golang:1.24

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o api ./cmd/api

EXPOSE 8080

CMD ["./api"] 