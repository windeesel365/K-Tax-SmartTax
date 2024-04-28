FROM golang:1.21 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./
COPY .env ./

RUN CGO_ENABLED=0 GOOS=linux go build -v -o myapp

FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /app

# copy built binary และ .env file
COPY --from=builder /app/myapp .
COPY --from=builder /app/.env .

# expose port 8080 to outside world
EXPOSE 8080

CMD ["./myapp"]
