FROM golang:alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN cd /app/cmd/cli && go build -o main .
FROM alpine:latest
COPY --from=build /app/cmd/cli/main /app/main
RUN apk --no-cache add ca-certificates && chmod +x /app/main
CMD ["/app/main"]