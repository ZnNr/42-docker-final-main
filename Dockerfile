FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod download

COPY *.go *.db ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /docker-final

CMD ["/docker-final"]