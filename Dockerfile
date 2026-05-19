FROM golang:1.24
LABEL authors="bikopougala"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Define a build argument
ARG MONGO_CONNECTION_URL

# Set the environment variable using the build argument
ENV MONGO_CONNECTION_URL=${MONGO_CONNECTION_URL}

COPY . .

EXPOSE 8080

CMD ["go", "run", "server.go"]