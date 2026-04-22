FROM golang:1.25.7 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o carrier ./cmd/carrier
RUN CGO_ENABLED=1 GOOS=linux go build -o probe ./cmd/probe

FROM gcr.io/distroless/base-debian12
WORKDIR /

COPY --from=builder /app/carrier /carrier
COPY --from=builder /app/probe /probe

# TODO: switch carrier to non-root (UserID/GroupID 65532) once the volume is
# initialized with the correct ownership for write access.
# USER 65532:65532

ENTRYPOINT ["/carrier"]