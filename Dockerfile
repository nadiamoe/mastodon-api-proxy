FROM golang:1.23-alpine3.19@sha256:5f3336882ad15d10ac1b59fbaba7cb84c35d4623774198b36ae60edeba45fd84 as builder

WORKDIR /proxy
COPY . .
RUN go build -o /bin/proxy .

FROM alpine:3.24.3@sha256:1b7232cda4ca3dc20b5d1247bdc9bfb631f2661d15d323bdd6fccc035ed1a9e7
COPY --from=builder /bin/proxy /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/proxy" ]
