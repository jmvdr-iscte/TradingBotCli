FROM golang:1.21.3

WORKDIR /usr/src/app

COPY . .

# RUN apt-get update && \
#    apt-get install -y golang-golang-x-tools && \
#    rm -rf /var/lib/apt/lists/*
   
RUN go mod tidy

