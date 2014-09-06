# DOCKER-VERSION 1.0.0
FROM debian:wheezy
MAINTAINER Amin Jams "aminjam@outlook.com"
RUN apt-get update

################## BEGIN INSTALLATION ######################
# Install MongoDB Following the Instructions at MongoDB Docs
# Ref: http://docs.mongodb.org/manual/tutorial/install-mongodb-on-ubuntu/

# Add the package verification key
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv 7F0CEB10

# Add MongoDB to the repository sources list
RUN echo 'deb http://downloads-distro.mongodb.org/repo/debian-sysvinit dist 10gen' | tee /etc/apt/sources.list.d/mongodb.list

# Update the repository sources list once more
RUN apt-get update

# Install MongoDB package (.deb)
RUN apt-get install -qy adduser mongodb-10gen

# Create the default data directory

# Install Supervisor
RUN apt-get -qy install supervisor

# Add configurations
ADD supervisord-app.conf /etc/supervisor/conf.d/supervisord-app.conf
ADD run.sh /run.sh
RUN chmod +x /*.sh
##################### INSTALLATION END #####################

# Expose the default port
RUN mkdir -p /home/app
EXPOSE 27017
WORKDIR /home/app
