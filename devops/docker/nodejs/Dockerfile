# DOCKER-VERSION 1.2.0
FROM ubuntu:latest
MAINTAINER Amin Jams "aminjam@outlook.com"

RUN apt-get update && \
  apt-get -y upgrade && \
  apt-get autoremove && \
  apt-get autoclean && \
  apt-get clean
RUN apt-get install -qy curl git
RUN curl -qsL https://deb.nodesource.com/setup | sudo bash -
RUN apt-get install -qy nodejs
RUN npm install -g bower grunt-cli

# Install Supervisor
RUN apt-get -qy install supervisor

# Add configurations
ADD supervisord-app.conf /etc/supervisor/conf.d/supervisord-app.conf
ADD run.sh /run.sh
ADD start-app.sh /start-app.sh
RUN chmod +x /*.sh

##################### INSTALLATION END #####################

RUN echo 'node:' `node --version` 'npm' `npm --version`
EXPOSE 3001
WORKDIR /home/app
