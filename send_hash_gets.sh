#!/bin/bash

# set -x
for i in {1..20}
do
  echo "calling fetch hash with value: "$i
  curl http://localhost:8080/hash/$i
  sleep .5
done
