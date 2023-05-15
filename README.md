# Kayvee

A distributed in-memory key-value store built using [hashicorp/memberlist](https://github.com/hashicorp/memberlist) with HTTP API

## Install

```shell
go get github.com/asim/kayvee
```

## Usage

```shell
kayvee
-nodes="": comma seperated list of nodes
-address=:4001 http server host:port
```

### Create Cluster

Start first node
```shell
kayvee
```

Make a note of the local node address
```
Local node 192.168.1.64:60496
Listening on :4001
```

Start second node with first node as part of the nodes list
```shell
kayvee --nodes=192.168.1.64:60496 --address=:4002
```

You should see the output
```
2015/10/17 22:13:49 [DEBUG] memberlist: Initiating push/pull sync with: 192.168.1.64:60496
Local node 192.168.1.64:60499
Listening on :4002
```

First node output will log the new connection
```shell
2015/10/17 22:13:49 [DEBUG] memberlist: TCP connection from: 192.168.1.64:60500
2015/10/17 22:13:52 [DEBUG] memberlist: Initiating push/pull sync with: 192.168.1.64:60499
```

## HTTP API

- **/get** - get a value
- **/set** - set a value
- **/del** - delete a value

Query params expected are `key` and `val`

```shell
# add
curl "http://localhost:4001/set?key=foo&val=bar"

# get
curl "http://localhost:4001/get?key=foo"

# delete
curl "http://localhost:4001/del?key=foo"
```

## HTTP UI

Browse to `localhost:4001` for a simple web form
