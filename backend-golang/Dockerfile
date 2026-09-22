FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -o /out/import ./cmd/import

FROM alpine:3

RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app

USER app

COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/import /usr/local/bin/import

EXPOSE 8082

CMD ["api"]
