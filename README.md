# go-http-server
A demo Golang http server example.

#### build
```shell
go build *.go
```

#### run
```shell
go run *.go
```

`curl --data "password=something" http://localhost:8080/hash`
Returns request number.

`curl http://localhost:8080/entries`
Returns POST requests made to "/hash" which successfully returned a request number
This is just for debugging

`curl http://localhost:8080/hash/req_num`
Returns SHA512 of the param `password` provided to "/hash" 5 seconds after
the request was made.

Once the server is up and running
```shell
./send_hash_posts.sh # runs a script to send 20 requests to the server
./send_hash_gets.sh # runs a script to output values for ids 1-20 and outputs message from server
```
Thanks to ample number of golang posts on handling http requests along with
doing regex matches for example:
https://github.com/benhoyt/go-routing/blob/master/match/route.go
