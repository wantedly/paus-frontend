FROM alpine:3.3
MAINTAINER Daisuke Fujita <dtanshi45@gmail.com> (@dtan4)

ENV DOCKER_COMPOSE_VERSION 1.6.0
ENV GLIBC_VERSION 2.22-r8
ENV GLIDE_VERSION 0.8.3
ENV ETCD_VERSION 2.2.5

RUN apk --update add ca-certificates docker && \
    wget -O /usr/local/bin/docker-compose https://github.com/docker/compose/releases/download/$DOCKER_COMPOSE_VERSION/docker-compose-Linux-x86_64 && \
    chmod +x /usr/local/bin/docker-compose && \
    wget -O glibc.apk https://github.com/andyshinn/alpine-pkg-glibc/releases/download/2.22-r8/glibc-$GLIBC_VERSION.apk && \
    wget -O glibc-bin.apk https://github.com/andyshinn/alpine-pkg-glibc/releases/download/2.22-r8/glibc-bin-$GLIBC_VERSION.apk && \
    apk add --allow-untrusted glibc.apk glibc-bin.apk && \
    /usr/glibc-compat/sbin/ldconfig /usr/glibc-compat/lib && \
    rm glibc.apk glibc-bin.apk && \
    apk del --purge ca-certificates && \
    rm -rf /var/cache/apk/*

RUN wget -qO /tmp/etcd-v$ETCD_VERSION-linux-amd64.tar.gz https://github.com/coreos/etcd/releases/download/v$ETCD_VERSION/etcd-v$ETCD_VERSION-linux-amd64.tar.gz && \
    cd /tmp && \
    tar zxf etcd-v$ETCD_VERSION-linux-amd64.tar.gz && \
    cp etcd-v$ETCD_VERSION-linux-amd64/etcdctl /usr/local/bin/etcdctl && \
    rm -rf etcd-v$ETCD_VERSION-linux-amd64.tar.gz etcd-v$ETCD_VERSION-linux-amd64

COPY . /go/src/github.com/dtan4/paus
RUN apk --update add git go make mercurial && \
    wget -O /tmp/glide.tar.gz https://github.com/Masterminds/glide/releases/download/0.8.3/glide-$GLIDE_VERSION-linux-amd64.tar.gz && \
    cd /tmp && \
    tar zxf glide.tar.gz && \
    cp linux-amd64/glide /usr/local/bin && \
    export GOPATH=/go && \
    export GO15VENDOREXPERIMENT=1 && \
    cd /go/src/github.com/dtan4/paus && \
    glide install && \
    make build && \
    mkdir /app && \
    cp bin/paus-frontend /app/paus-frontend && \
    cp -R templates /app/templates && \
    cd /app && \
    rm -rf /go \
           /usr/local/bin/glide \
           /tmp/glide.tar.gz \
           /tmp/linux-amd64 && \
    apk del --purge git go make mercurial && \
    rm -rf /var/cache/apk/*

WORKDIR /app
EXPOSE 8080

CMD ["./paus-frontend"]
