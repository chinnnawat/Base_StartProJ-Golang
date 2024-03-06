# use official Golang image
FROM golang:1.21.6

# set working directory
WORKDIR /usr/src/app

RUN go install github.com/cosmtrek/air@latest

COPY . .
