FROM golang:1.23
LABEL authors="bikopougala"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "server.go"]