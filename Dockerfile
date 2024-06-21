FROM golang:1.22.1 as base
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o transip-ddns cmd/transip-ddns/main.go
RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  user

FROM scratch
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

COPY --from=base /app/transip-ddns .

USER user:user
ENTRYPOINT ["/transip-ddns"]
