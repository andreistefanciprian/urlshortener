FROM golang:1.26.2 AS build

WORKDIR /app

COPY go.mod ./

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o frontend2 ./cmd

FROM gcr.io/distroless/static-debian12:nonroot@sha256:e8a4044e0b4ae4257efa45fc026c0bc30ad320d43bd4c1a7d5271bd241e386d0

WORKDIR /app

COPY --from=build /app/frontend2 .
COPY --from=build /app/templates ./templates

USER nonroot:nonroot

ENTRYPOINT ["/app/frontend2"]
