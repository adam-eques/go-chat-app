FROM golang:1.17

WORKDIR /chat-app
COPY go.mod .
COPY go.sum .
RUN go mod download

RUN apt-get -y update
COPY . .

RUN go build -o chat-app

CMD ["./chat-app"]

