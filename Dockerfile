FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /launcher-server .

FROM alpine:3.19 AS production

RUN apk --no-cache add ca-certificates
COPY --from=build /launcher-server /launcher-server

EXPOSE 8080

CMD ["/launcher-server"]
