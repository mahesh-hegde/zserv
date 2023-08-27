ARG BASE_VERSION=latest
FROM golang:1.20 as build

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o /go/bin/zserv

FROM gcr.io/distroless/static-debian11:${BASE_VERSION}
COPY --from=build /go/bin/zserv /
WORKDIR /app/
ENTRYPOINT ["/zserv"]
LABEL org.opencontainers.image.source=https://github.com/mahesh-hegde/zserv
LABEL org.opencontainers.image.description="A small local HTTP server to serve from ZIP files without unpacking them"
## docker run -v $PWD:/app -u $(id -u):$(id -g) ...