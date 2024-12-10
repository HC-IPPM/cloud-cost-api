FROM golang:1.23.3 as builder
ARG CGO_ENABLED=0
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o ./cloud-cost-api

# Stage 2
FROM gcr.io/distroless/static-debian12
# FROM alpine:3.14

WORKDIR /app
COPY --from=builder /build/cloud-cost-api ./cloud-cost-api
COPY --from=builder /build/service_account.json ./service_account.json
CMD ["/app/cloud-cost-api"]
