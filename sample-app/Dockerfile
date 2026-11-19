FROM golang:1.27.4-alpine3.24 AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sample-app .

FROM alpine:3.24.3
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/sample-app /usr/local/bin/sample-app
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/sample-app"]

