## Multistage build
FROM golang:1.17-alpine as build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src
COPY . .
RUN go mod download
RUN go build -o /ext-proc

## Multistage deploy
FROM golang:1.17-alpine

WORKDIR /
COPY --from=build /ext-proc /ext-proc

ENTRYPOINT ["/ext-proc"]
