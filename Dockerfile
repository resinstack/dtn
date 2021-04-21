FROM golang:1.16-alpine AS build
WORKDIR /build
COPY . .
RUN go mod vendor && CGO_ENABLED=no go build -tags netgo -ldflags '-w' -o /dtn .

FROM scratch
LABEL org.opencontainers.image.source https://github.com/the-maldridge/dtn
COPY --from=build /dtn /dtn
ENTRYPOINT ["/dtn"]
