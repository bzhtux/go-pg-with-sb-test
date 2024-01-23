# builder
FROM golang:alpine AS build-env
LABEL maintainer="Yannick Foeillet <bzhtux@gmail.com>"

# wokeignore:rule=he/him/his
RUN apk --no-cache add build-base git mercurial gcc curl
RUN mkdir -p /go/src/github.com/bzhtux/go-pg-with-sb-test
ADD . /go/src/github.com/bzhtux/go-pg-with-sb-test
# RUN cd /go/src/github.com/bzhtux/go-pg-with-sb-test && go get ./... && go build -o gopg main.go
WORKDIR /go/src/github.com/bzhtux/go-pg-with-sb-test
RUN go get ./...
RUN go build -o gopg main.go


# final image
FROM alpine
LABEL maintainer="Yannick Foeillet <bzhtux@gmail.com>"

# wokeignore:rule=he/him/his
RUN apk --no-cache add curl jq
RUN adduser -h /app -s /bin/sh -u 1000 -D app
WORKDIR /app
COPY --from=build-env /go/src/github.com/bzhtux/go-pg-with-sb-test/gopg /app/
USER 1000
ENTRYPOINT ./gopg