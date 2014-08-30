# DOCKER-VERSION 1.2.0
FROM debian:wheezy
MAINTAINER Amin Jams "aminjam@outlook.com"

RUN apt-get update && \
  apt-get -y upgrade && \
  apt-get autoremove && \
  apt-get autoclean && \
  apt-get clean
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys ABF5BD827BD9BF62
RUN echo "deb http://www.nginx.org/packages/debian/ wheezy nginx" > /etc/apt/sources.list.d/nginx.list
RUN apt-get update
RUN apt-get install -y nginx

RUN rm /etc/nginx/nginx.conf /etc/nginx/mime.types
ADD nginx.conf /etc/nginx/nginx.conf
ADD mime.types /etc/nginx/mime.types
RUN mkdir /etc/nginx/ssl
ADD default /etc/nginx/sites-enabled/default
ADD default-ssl /etc/nginx/sites-available/default-ssl
RUN mkdir -p /home/app
##################### INSTALLATION END #####################

EXPOSE 80 443
WORKDIR /home/app
CMD ["nginx"]
