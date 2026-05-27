FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/echallan-calculator main.go

FROM alpine:3.18

WORKDIR /opt/egov

COPY --from=builder /app/echallan-calculator /opt/egov/echallan-calculator

EXPOSE 8078

CMD ["/opt/egov/echallan-calculator"]

