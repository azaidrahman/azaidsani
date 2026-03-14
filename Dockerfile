FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod main.go ./
RUN go build -o server .

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/server .
COPY templates/ ./templates/
COPY static/ ./static/
CMD ["./server"]
