[![Go Report Card](https://goreportcard.com/badge/github.com/peerbridge/peerbridge)](https://goreportcard.com/report/github.com/peerbridge/peerbridge)

Generate a new [Block](https://github.com/peerbridge/peerbridge/blob/master/pkg/block/block.go) 
given by using [cURL](https://curl.haxx.se/) on the command line.

```bash
$ curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"Index":1,"Timestamp":"2020-06-11T10:30:45Z","PrevHash":"c267f24d15144548de1d0f6097a5e7e040614fa28259474511d6e7691508d75b"}' \
  http://localhost:8000/block/new
```

Generate the sha256 hash value for a [Block](https://github.com/peerbridge/peerbridge/blob/master/pkg/block/block.go) 
given as json data by using [cURL](https://curl.haxx.se/) on the command line.

```bash
$ curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"Index":1,"Timestamp":"2020-06-11T10:30:45Z","PrevHash":"c267f24d15144548de1d0f6097a5e7e040614fa28259474511d6e7691508d75b"}' \
  http://localhost:8000/block/hash
```