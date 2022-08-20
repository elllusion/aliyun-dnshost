FROM golang:1.17.4-alpine as mybuilder
WORKDIR $GOPATH/src/github.com/aliyun-dnshost
ADD . $GOPATH/src/github.com/aliyun-dnshost
COPY config.yaml /etc/alidnshost-config.yaml
RUN go build .
RUN cp -f ./aliyun-dnshost /usr/local/bin/aliyun-dnshost

ENTRYPOINT  ["./aliyun-dnshost"]
CMD ["-c", "/etc/alidnshost-config.yaml"]

FROM alpine:3.16.0
COPY --from=mybuilder /usr/local/bin/aliyun-dnshost /usr/local/bin/aliyun-dnshost
COPY --from=mybuilder /etc/alidnshost-config.yaml /etc/alidnshost-config.yaml
ENTRYPOINT  ["/usr/local/bin/aliyun-dnshost"]
CMD ["-c", "/etc/alidnshost-config.yaml"]

