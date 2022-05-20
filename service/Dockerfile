FROM golang:1.17-alpine as build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /src
COPY . .
RUN go mod download
RUN go build -o /service main.go

FROM envoyproxy/envoy-alpine:v1.19-latest
COPY --from=build /src /src
COPY --from=build /service /service
CMD ["/bin/sh", "/src/start_service.sh"]
