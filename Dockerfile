# builder
# FROM harbor.h2o-2-22354.h2o.vmware.com/library/golang:alpine AS build-env
# FROM golang:alpine AS build-env
FROM ghcr.io/bzhtux/golang:latest AS build-env
LABEL maintainer="Yannick Foeillet <bzhtux@gmail.com>"

WORKDIR /app

# wokeignore:rule=he/him/his
# RUN apk --no-cache add build-base git mercurial gcc curl
# RUN mkdir -p /go/src/github.com/bzhtux/go-pg-with-sb-test
# ADD . /go/src/github.com/bzhtux/go-pg-with-sb-test
ADD go.mod go.sum ./
RUN go mod download

COPY *.go ./
# RUN cd /go/src/github.com/bzhtux/go-pg-with-sb-test && go get ./... && go build -o gopg main.go
# WORKDIR /go/src/github.com/bzhtux/go-pg-with-sb-test
# RUN go get ./...
# RUN go build -o gopg main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /gopg

# final image
# FROM harbor.h2o-2-22354.h2o.vmware.com/library/alpine
FROM scratch
LABEL maintainer="Yannick Foeillet <bzhtux@gmail.com>"

# wokeignore:rule=he/him/his
# RUN apk --no-cache add curl jq
# RUN adduser -h /app -s /bin/sh -u 1000 -D app
WORKDIR /app
COPY --from=build-env /gopg /app/
# USER 1000

EXPOSE 8080

# ENTRYPOINT ./gopg

# Run
CMD ["/app/gopg"]

# ENTRYPOINT /app/gopg
