FROM golang:1.24

WORKDIR /build
COPY . ./
RUN CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o crontask ./

FROM docker:28.1.1-dind-alpine3.21
RUN apk add --no-cache tzdata ca-certificates curl
COPY --from=0 /build/crontask /bin/

ENTRYPOINT [ "/bin/crontask" ]
