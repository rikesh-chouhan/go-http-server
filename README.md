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
