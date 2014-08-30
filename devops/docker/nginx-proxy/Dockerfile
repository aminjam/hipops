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
RUN apt-get install -qy nginx wget

RUN echo "daemon off;" >> /etc/nginx/nginx.conf
#fix for long server names
RUN sed -i 's/# server_names_hash_bucket/server_names_hash_bucket/g' /etc/nginx/nginx.conf

RUN wget -q --no-check-certificate https://godist.herokuapp.com/projects/ddollar/forego/releases/current/linux-amd64/forego -P /usr/local/bin
RUN chmod u+x /usr/local/bin/forego

RUN wget -q --no-check-certificate https://github.com/jwilder/docker-gen/releases/download/0.3.3/docker-gen-linux-amd64-0.3.3.tar.gz -O docker-gen.tar.gz && \
  tar -xvzf docker-gen.tar.gz && \
  mv docker-gen /usr/local/bin && \
  chmod u+x /usr/local/bin/docker-gen &&\
  rm docker-gen.tar.gz

RUN mkdir /etc/nginx/sites-enabled
RUN mkdir /home/app
ADD . /home/app
RUN rm /etc/nginx/nginx.conf /etc/nginx/mime.types
ADD nginx.conf /etc/nginx/nginx.conf
ADD mime.types /etc/nginx/mime.types
ADD proxy_params /etc/nginx/proxy_params
##################### INSTALLATION END #####################

EXPOSE 80
ENV DOCKER_HOST unix:///tmp/docker.sock
WORKDIR /home/app
CMD ["forego", "start", "-r"]
