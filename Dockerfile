FROM golang:1.20.5-alpine3.18 AS builder
WORKDIR /

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
RUN apt install build-essential

FROM gcr.io/distroless/base-debian11:nonroot
WORKDIR /

COPY --from=builder /app ./
COPY --from=builder /templates ./templates
COPY --from=builder /uploads ./uploads

EXPOSE 8080

CMD ["./app"]