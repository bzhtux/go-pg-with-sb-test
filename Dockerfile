# builder
FROM ghcr.io/bzhtux/golang:latest AS build-env
LABEL maintainer="Yannick Foeillet <bzhtux@gmail.com>"

WORKDIR /app

ADD go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /gopg

# final image
FROM scratch
LABEL maintainer="Yannick Foeillet <bzhtux@gmail.com>"

WORKDIR /app
COPY --from=build-env /gopg /app/
#USER 1000

EXPOSE 8080

# Run
CMD ["/app/gopg"]
