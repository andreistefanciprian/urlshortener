FROM golang:1.25@sha256:8305f5fa8ea63c7b5bc85bd223ccc62941f852318ebfbd22f53bbd0b358c07e1 AS build

WORKDIR /app

COPY go.mod ./

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o url-read ./cmd

FROM gcr.io/distroless/static-debian12:nonroot@sha256:e8a4044e0b4ae4257efa45fc026c0bc30ad320d43bd4c1a7d5271bd241e386d0

WORKDIR /app

COPY --from=build /app/url-read .

# Use the nonroot user (UID 65532)
USER nonroot:nonroot

ENTRYPOINT ["/app/url-read"]