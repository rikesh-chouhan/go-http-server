#!/bin/bash

# set -x
pw="darksideofthemoon"
for i in {1..20}
do
  val="$pw$i"
  echo "calling with password value: "$val
  curl --data "password=$val" http://localhost:8080/hash
  sleep .5
done