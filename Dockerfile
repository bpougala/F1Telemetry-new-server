FROM golang:1.23
LABEL authors="bikopougala"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ENV MONGO_CONNECTION_URL="mongodb+srv://doapps-4cff79f9-cdec-4e80-8807-8e0f9194e32a:zBef9oM5461u8S27@db-mongodb-lon1-34655-f8b01560.mongo.ondigitalocean.com/admin?tls=true&authSource=admin&replicaSet=db-mongodb-lon1-34655"

COPY . .

EXPOSE 8080

CMD ["go", "run", "server.go"]