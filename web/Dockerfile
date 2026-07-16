FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk add --no-cache caddy && \
    rm -rf /var/cache/apk/*

RUN caddy version

RUN mkdir -p /var/www/html
COPY public/ /var/www/html/
COPY Caddyfile /etc/caddy/Caddyfile

RUN apk add --no-cache sudo libcap && \
    adduser -D -s /bin/sh admin && \
    echo "admin:admin" | chpasswd && \
    addgroup admin wheel && \
    echo "%wheel ALL=(ALL) ALL" > /etc/sudoers.d/wheel && \
    chmod 0440 /etc/sudoers.d/wheel && \
    passwd -l root && \
    setcap cap_net_bind_service=+ep /usr/sbin/caddy && \
    chown -R admin:admin /var/www/html /etc/caddy /var/log

ENV HOME=/home/admin

USER admin
EXPOSE 80 443

ENTRYPOINT ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]