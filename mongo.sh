#! /bin/bash

docker pull mongo:4.4

docker stop $(docker ps -aq)

docker rm $(docker ps -aq)

docker run -d -p 27017:27017 --name test-mongo -v data-vol:/data/db -e MONGODB_INITDB_ROOT_USERNAME=test -e MONGODB_INITDB_ROOT_PASSWORD=test123 mongo:4.4
