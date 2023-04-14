FROM golang:1.18

WORKDIR /build
COPY . ./
RUN CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o crontask ./

FROM alpine:3.12
RUN apk add --no-cache tzdata
COPY --from=0 /build/crontask /bin/

ENTRYPOINT [ "/bin/crontask" ]
