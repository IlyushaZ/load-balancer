#!/bin/bash

docker network create balancer

for i in {1..5} ; do
    docker container run --name static-$i -p 808$i:80 -d \
    --net balancer --net-alias static -v "$PWD"/templates:/usr/share/nginx/html nginx
done

docker container run --rm --net balancer alpine nslookup static

rx='([1-9]?[0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])'
raw=$(docker container run --rm --net balancer alpine nslookup static)

for ip in $raw; do
  if [[ ($ip =~ ^$rx\.$rx\.$rx\.$rx$) && ($ip != "192.168.0.1") ]] ; then
      echo "$ip" >> "backends.txt"
  fi
done

docker build -t ilyaz/load-balancer .
docker container run -p 8087:80 -d --net balancer ilyaz/load-balancer