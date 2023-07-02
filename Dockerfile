FROM golang:1.20.5-alpine3.18 AS builder
WORKDIR /

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=1 GOOS=linux go build -o /app -a -ldflags '-linkmode external -extldflags "-static"' .

FROM scratch
WORKDIR /

COPY --from=builder /app ./
COPY --from=builder /templates ./templates
COPY --from=builder /uploads ./uploads

EXPOSE 8080

CMD ["./app"]