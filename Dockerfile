FROM golang AS development

WORKDIR /app

COPY . .
RUN go mod download

CMD ["go", "run", "."]

FROM golang AS builder

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /tails-social-go

FROM alpine AS production

WORKDIR /

COPY --from=builder /tails-social-go /

CMD ["/tails-social-go"]