FROM golang:1.23-alpine3.19@sha256:5f3336882ad15d10ac1b59fbaba7cb84c35d4623774198b36ae60edeba45fd84 as builder

WORKDIR /proxy
COPY . .
RUN go build -o /bin/proxy .

FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11
COPY --from=builder /bin/proxy /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/proxy" ]
