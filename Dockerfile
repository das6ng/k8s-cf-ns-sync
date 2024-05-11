FROM golang:1.22-alpine AS build
WORKDIR /build-src
COPY . .
RUN go build -trimpath -o cf-ns-sync ./bin

FROM alpine
WORKDIR /app
COPY --from=build --chown=1000:1000 /build-src/cf-ns-sync .
ENTRYPOINT [ "/app/cf-ns-sync" ]
