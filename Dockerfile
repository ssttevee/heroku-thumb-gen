FROM golang:alpine
WORKDIR /workspace
COPY main.go .
RUN go build main.go

FROM alpine:latest
RUN apk add --update --no-cache imagemagick ghostscript-fonts
COPY --from=0 /workspace/main /usr/bin/serve
ENTRYPOINT /usr/bin/serve
