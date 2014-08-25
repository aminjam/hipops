# DOCKER-VERSION 1.1.2
FROM ubuntu:14.04
MAINTAINER Amin Jams "aminjam@outlook.com"

RUN apt-get update
RUN apt-get install -qy curl
RUN curl -sL https://deb.nodesource.com/setup | sudo bash -
RUN apt-get -q update
RUN apt-get install -qy nodejs
RUN npm cache clean

# Install Supervisor
RUN apt-get -qy install supervisor

# Add configurations
ADD supervisord-app.conf /etc/supervisor/conf.d/supervisord-app.conf
ADD run.sh /run.sh
ADD start.sh /start.sh
RUN chmod +x /*.sh

##################### INSTALLATION END #####################

EXPOSE 3001
WORKDIR /home/app
