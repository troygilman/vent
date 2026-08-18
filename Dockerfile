# Build and run the examples/basic SQLite admin app.
#   docker build -t vent-example .
#   docker run --rm -p 8080:8080 vent-example
# Then open http://localhost:8080/admin/ (admin@vent.com / test_user).

FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# go-sqlite3 requires CGO; the golang image already includes a C compiler.
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/vent-example ./examples/basic/cmd/server

# Same community Atlas CLI the repo's Cloud Agent install script uses.
ARG TARGETARCH
RUN curl -fsSL "https://atlasbinaries.com/atlas/atlas-community-linux-${TARGETARCH}-latest" -o /out/atlas \
    && chmod +x /out/atlas

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 65532 --create-home --home-dir /app vent

COPY --from=builder /out/atlas /usr/local/bin/atlas
COPY --from=builder /out/vent-example /usr/local/bin/vent-example
COPY --chown=vent:vent examples/basic/ent/migrate/migrations /app/examples/basic/ent/migrate/migrations
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

WORKDIR /app
RUN chmod +x /usr/local/bin/docker-entrypoint.sh \
    && mkdir -p /app/tmp \
    && chown -R vent:vent /app

USER vent
EXPOSE 8080

ENTRYPOINT ["docker-entrypoint.sh"]
