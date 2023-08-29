FROM golang:1.20 as builder

# Build arguments for this image (used as -X args in ldflags)
ARG VAULTPAL_VERSION=""
ARG VAULTPAL_COMMIT=""

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build \
  -ldflags="-w -s \
  -X 'github.com/dbschenker/vaultpal/cmd.Version=${VAULTPAL_VERSION}' \
  -X 'github.com/dbschenker/vaultpal/cmd.Commit=${VAULTPAL_COMMIT}' \
  -extldflags '-static'" \
  -a -o main .

FROM gcr.io/distroless/static
COPY --from=builder /app/main /vaultpal
ENTRYPOINT ["/vaultpal"]
