FROM golang:1.26.3 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o mothership ./cmd/mothership
RUN CGO_ENABLED=1 GOOS=linux go build -o probe ./cmd/probe

FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /app/mothership /app/mothership
COPY --from=builder /app/probe /app/probe
COPY --from=builder /app/sql/schema /app/sql/schema

# TODO: switch mothership to non-root (UID/GID 65532) once the volume is
# initialized with the correct ownership for write access.
# USER 65532:65532

ENTRYPOINT ["/app/mothership"]