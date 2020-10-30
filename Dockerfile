FROM golang:1.15 as builder

# Build go binary  
WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY ./src/*go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /go/bin/app
COPY --from=builder /go/src/app/app .

COPY config.yml .

EXPOSE 9000

CMD ["./app"]  