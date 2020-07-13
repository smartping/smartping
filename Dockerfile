FROM golang:1.13.0 AS builder
COPY . /data/src/smartping
WORKDIR /data/src/smartping/bin
ENV GOPROXY "https://goproxy.cn,direct"
ENV CGO_ENABLED 1
RUN go build -mod=vendor -ldflags "-linkmode external -extldflags -static" -o smartping-linux ../src/smartping.go

FROM centos:centos7
ENV CONSUL_ENDPOINT "10.93.10.66:80"
EXPOSE 8899/tcp
COPY --from=builder /data/src/smartping/bin/smartping-linux /data/smartping/bin/smartping
COPY --from=builder /data/src/smartping/conf/config-base.json /data/smartping/conf/config-base.json
COPY --from=builder /data/src/smartping/conf/seelog.xml /data/smartping/conf/seelog.xml
COPY --from=builder /data/src/smartping/html /data/smartping/html
COPY --from=builder /data/src/smartping/db/database-base.db /data/smartping/db/database-base.db
COPY --from=builder /data/src/smartping/conf/seelog.xml /data/smartping/conf/seelog.xml
#RUN apk --no-cache add ca-certificates && mkdir /data/smartping/db
# RUN mkdir /data/smartping/db

WORKDIR /data/smartping/bin

CMD ["./smartping"]