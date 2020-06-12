[![Go Report Card](https://goreportcard.com/badge/github.com/peerbridge/peerbridge)](https://goreportcard.com/report/github.com/peerbridge/peerbridge)

Generate a new [Block](https://github.com/peerbridge/peerbridge/blob/master/pkg/block/block.go) 
given by using [cURL](https://curl.haxx.se/) on the command line.

```bash
$ curl --header "Content-Type: application/json" \
  --data '{"index":1,"timestamp":"2020-06-11T10:30:45Z","prev":"c267f24d15144548de1d0f6097a5e7e040614fa28259474511d6e7691508d75b"}' \
  http://localhost:8000/block/new
```

Generate the sha256 hash value for a [Block](https://github.com/peerbridge/peerbridge/blob/master/pkg/block/block.go) 
given as json data by using [cURL](https://curl.haxx.se/) on the command line.

```bash
$ curl --header "Content-Type: application/json" \
  --data '{"index":1,"timestamp":"2020-06-11T10:30:45Z","prev":"c267f24d15144548de1d0f6097a5e7e040614fa28259474511d6e7691508d75b"}' \
  http://localhost:8000/block/hash
```

### Documentation

Launch a local http server with the code documentation using

```bash
$ godoc -http=:6060
```

Then navigate to [localhost:6060/pkg/github.com/peerbridge/peerbridge/](http://localhost:6060/pkg/github.com/peerbridge/peerbridge/)
