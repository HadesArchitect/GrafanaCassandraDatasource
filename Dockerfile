FROM node:11

RUN apt-get -qq update
RUN apt-get install -y chromium

RUN mkdir -p /opt/karma
WORKDIR /opt/karma
RUN npm install -g karma
CMD karma start
