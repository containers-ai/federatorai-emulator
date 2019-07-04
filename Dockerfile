# Build the manager binary
FROM golang:1.11.5-stretch as builder

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Copy in the go src
WORKDIR /go/src/github.com/containers-ai/federatorai-emulator
ADD . .

# Build
RUN make build

# Prepare the package into a thin image
FROM alpine:latest
# FROM busybox:latest

ENV AIHOME=/opt/alameda/federatorai-emulator \
    USER_UID=1001 \
    USER_NAME=alameda

WORKDIR ${AIHOME}
COPY --from=builder /go/src/github.com/containers-ai/federatorai-emulator/tini /sbin/tini
COPY --from=builder /go/src/github.com/containers-ai/federatorai-emulator/install_root.tgz /tmp/

RUN set -x \
    # && apt-get update && apt-get install -y --force-yes --no-install-recommends vim logrotate \
    # && apt-get autoclean && apt-get autoremove && rm -rf /var/lib/apt/lists/* \
    && echo "${USER_NAME}:x:${USER_UID}:0:Federator.ai:${AIHOME}:/bin/sh" >> /etc/passwd \
    # The following lines for logrotate - startup script will add running user id into /etc/passwd
    && chmod g+w /etc/passwd \
    && mkdir -p /var/lib/logrotate /var/log \
    # logrotate need writable permission on "/var/lib/logrotate"
    && chmod ug+w /var/lib/logrotate /var/log \
    # install packages
    && cd / && tar xzvf /tmp/install_root.tgz && rm -fv /tmp/install_root.tgz \
    && chown -R ${USER_UID}:root ${AIHOME} && chmod -R ug+w ${AIHOME} \
    && cd /etc/ \
    && ln -s /opt/alameda/federatorai-emulator/etc/logrotate.conf logrotate.conf \
    && sed -i -e '/su root syslog/d' /etc/logrotate.conf \
    && mkdir -pv -m 775 /var/log/alameda && chown -R ${USER_UID}:root /var/log/alameda

USER ${USER_UID}
ENTRYPOINT ["/init.sh"]

