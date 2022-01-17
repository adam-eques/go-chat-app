FROM golang:1.17

WORKDIR /home/acentior/Desktop/Go/chat-app
COPY . .

RUN apt-get -y update
RUN go install -v ./...
RUN go build -o chat-app

CMD ./chat-app

