# DOCKER-VERSION 1.2.0
FROM debian:wheezy
MAINTAINER Amin Jams "aminjam@outlook.com"

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update -qy && \
    apt-get install -qy socat supervisor && \
    apt-get clean

ADD start-socat.sh /start-socat.sh
ADD start.sh /start.sh

ADD supervisord-socat.conf /etc/supervisor/conf.d/supervisord-socat.conf
RUN chmod 755 /*.sh
