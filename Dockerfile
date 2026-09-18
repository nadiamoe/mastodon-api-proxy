FROM golang:1.23-alpine3.19@sha256:5f3336882ad15d10ac1b59fbaba7cb84c35d4623774198b36ae60edeba45fd84 as builder

WORKDIR /proxy
COPY . .
RUN go build -o /bin/proxy .

FROM alpine:3.24.2@sha256:3cf95fe0816180395592b8373f3ec60663f076127617bbacb4eacf9667afe2e9
COPY --from=builder /bin/proxy /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/proxy" ]
