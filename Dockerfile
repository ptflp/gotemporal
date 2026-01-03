FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/app ./cmd/app

FROM gcr.io/distroless/base-debian12

COPY --from=builder /bin/app /bin/app

EXPOSE 3000

ENTRYPOINT ["/bin/app"]

