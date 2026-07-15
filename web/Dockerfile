FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk add --no-cache caddy && \
    rm -rf /var/cache/apk/*

RUN caddy version

RUN mkdir -p /var/www/html

COPY public/ /var/www/html/

COPY Caddyfile /etc/caddy/Caddyfile

RUN apk add --no-cache sudo su-exec libcap

RUN adduser -D -s /bin/sh admin && \
    echo "admin:admin" | chpasswd

RUN adduser admin wheel && \
    echo "%wheel ALL=(ALL) ALL" > /etc/sudoers.d/wheel && \
    chmod 0440 /etc/sudoers.d/wheel

RUN passwd -l root

RUN setcap cap_net_bind_service=+ep /usr/sbin/caddy

RUN chown -R admin:admin /var/www/html /etc/caddy /var/log

USER admin
EXPOSE 80 443

ENTRYPOINT ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
