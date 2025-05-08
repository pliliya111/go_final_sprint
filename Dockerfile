FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

EXPOSE ${port_a}
EXPOSE ${port_o}
EXPOSE 8080

COPY ./internal ./internal
COPY ./cmd ./cmd