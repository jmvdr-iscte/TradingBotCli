FROM golang:1.21.3

WORKDIR /usr/src/app

COPY . .
RUN go mod tidy