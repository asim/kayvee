# Memberlist

A basic http key/value example of how to use [hashicorp/memberlist](https://github.com/hashicorp/memberlist)

### Install
```shell
$ go get github.com/asim/memberlist
```

### Usage
```shell
$ memberlist
Usage of /Users/asim/checkouts/bin/memberlist:
  -members="": comma seperated list of members
  -port=4001: http port
```

### Run
Start first node
```shell
$ memberlist
Local member 192.168.1.64:60496
Listening on :4001
```

Start second node
```shell
$ memberlist --members=192.168.1.64:60496 --port=4002
2015/10/17 22:13:49 [DEBUG] memberlist: Initiating push/pull sync with: 192.168.1.64:60496
Local member 192.168.1.64:60499
Listening on :4002
```

First node output
```shell
2015/10/17 22:13:49 [DEBUG] memberlist: TCP connection from: 192.168.1.64:60500
2015/10/17 22:13:52 [DEBUG] memberlist: Initiating push/pull sync with: 192.168.1.64:60499
```

### Key Value

add
```shell
$ curl "http://localhost:4001/add?key=foo&val=bar"
```

get
```shell
$ curl "http://localhost:4001/get?key=foo"
```

del
```shell
$ curl "http://localhost:4001/del?key=foo"
```
