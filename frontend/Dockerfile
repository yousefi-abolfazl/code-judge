FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 2020

CMD ["go", "run", "serve.go"] 