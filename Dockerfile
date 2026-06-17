FROM golang:1.23-alpine3.19@sha256:5f3336882ad15d10ac1b59fbaba7cb84c35d4623774198b36ae60edeba45fd84 as builder

WORKDIR /proxy
COPY . .
RUN go build -o /bin/proxy .

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
COPY --from=builder /bin/proxy /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/proxy" ]
