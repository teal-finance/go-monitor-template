# builder image
FROM golang:1.15 as builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp main.go

# final image
FROM alpine:latest
COPY --from=builder /build/myapp .
EXPOSE 8080
ENTRYPOINT ["./myapp"]
