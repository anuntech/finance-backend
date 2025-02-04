FROM golang:1.22.5
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -a . .
CMD ./api-willchat-golang