[https://peerbridge.herokuapp.com](https://peerbridge.herokuapp.com/)

**NOTICE**

UNDER DEVELOPMENT

### Quick Start (Docker)

Start the PeerBridge Blockchain Server including PostgreSQL Database and Adminer using

```bash
$ docker-compose up -d
```

### Quick Start (Development)

Start the PostgreSQL Database and Adminer using
```bash
$ docker-compose up -d database adminer
```

Then, start the PeerBridge Blockchain Server
```bash
$ go run main.go
```

### Endpoints

##### Get peer to peer multiaddresses

Request:
```bash
$ curl localhost:8080/blockchain/peer/urls
```

Response:
```json
[
    {
        "Scheme":"",
        "Opaque":"",
        "User":null,
        "Host":"",
        "Path":"/ip6/::1/tcp/52822/p2p/QmQ6zsaeVWTm7XxBXuzKHQtBoNxPpdCSyA3HjDRL9zVnaH",
        "RawPath":"",
        "ForceQuery":false,
        "RawQuery":"",
        "Fragment":""
    },
    {
        "Scheme":"",
        "Opaque":"",
        "User":null,
        "Host":"",
        "Path":"/ip4/192.168.178.23/tcp/52821/p2p/QmQ6zsaeVWTm7XxBXuzKHQtBoNxPpdCSyA3HjDRL9zVnaH",
        "RawPath":"",
        "ForceQuery":false,
        "RawQuery":"",
        "Fragment":""
    },
    {
        "Scheme":"",
        "Opaque":"",
        "User":null,
        "Host":"",
        "Path":"/ip4/127.0.0.1/tcp/52821/p2p/QmQ6zsaeVWTm7XxBXuzKHQtBoNxPpdCSyA3HjDRL9zVnaH",
        "RawPath":"",
        "ForceQuery":false,
        "RawQuery":"",
        "Fragment":""
    }
]
```

##### Get a transaction by its index

Request:
```bash
$ curl localhost:8080/blockchain/transactions?index=9aa55d17-8d4b-41da-8512-d89c1b8e8bc8
```

Response:
```json
{
    "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
    "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "timestamp": "2020-10-03T09:13:37.26202Z",
    "data": "SW5jcm95YWJsZQ==",
    "blockIndex": "fb29934d-a688-4508-bfb6-29de21ac54d7"
}
```

#### Create a new transaction

Request:
```bash
$ curl --header "Content-Type: application/json" http://localhost:8080/blockchain/transactions/new --data
{
    "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "data": "SW5jcm95YWJsZQ=="
}
```

Response:
```json
{
    "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
    "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "timestamp": "2020-10-03T09:13:37.26202Z",
    "data": "SW5jcm95YWJsZQ==",
    "blockIndex": ""
}
```

#### Get all transaction for a given public key

Request:
```bash
$ curl --header "Content-Type: application/json" http://localhost:8080/blockchain/transactions/filter --data
{
    "publicKey": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n"
}
```

Response:
```json
[
    {
        "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
        "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "timestamp": "2020-10-03T09:13:37.26202Z",
        "data": "SW5jcm95YWJsZQ==",
        "blockIndex": "fb29934d-a688-4508-bfb6-29de21ac54d7"
    }
]
```

#### Get all transaction for a given public key after a certain timestamp

Request:
```bash
$ curl --header "Content-Type: application/json" http://localhost:8080/blockchain/transactions/filter --data
{
    "publicKey": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "timestamp": "2020-10-03T09:10:10Z"

}
```

Response:
```json
[
    {
        "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
        "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "timestamp": "2020-10-03T09:13:37.26202Z",
        "data": "SW5jcm95YWJsZQ==",
        "blockIndex": "fb29934d-a688-4508-bfb6-29de21ac54d7"
    }
]
```

#### Get all transactions received by a given public key

```bash
$ curl --header "Content-Type: application/json" http://localhost:8080/blockchain/transactions/filter --data
{
    "publicKey": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n"
}
```

Response:
```json
[
    {
        "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
        "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "timestamp": "2020-10-03T09:13:37.26202Z",
        "data": "SW5jcm95YWJsZQ==",
        "blockIndex": "fb29934d-a688-4508-bfb6-29de21ac54d7"
    }
]
```

#### Get all transactions received by a given public key after a certain timestamp

```bash
$ curl --header "Content-Type: application/json" http://localhost:8080/blockchain/transactions/filter --data
{
    "publicKey": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
    "timestamp": "2020-10-03T09:10:10Z"
}
```

Response:
```json
[
    {
        "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
        "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
        "timestamp": "2020-10-03T09:13:37.26202Z",
        "data": "SW5jcm95YWJsZQ==",
        "blockIndex": "fb29934d-a688-4508-bfb6-29de21ac54d7"
    }
]
```

##### Get a block by its index

```bash
$ curl http://localhost:8080/blockchain/blocks?index=fb29934d-a688-4508-bfb6-29de21ac54d7
```

Response:
```json
{
    "index": "fb29934d-a688-4508-bfb6-29de21ac54d7",
    "timestamp": "2020-10-03T09:13:40.189489Z",
    "transactions": [
        {
            "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
            "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
            "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
            "timestamp": "2020-10-03T09:13:37.26202Z",
            "data": "SW5jcm95YWJsZQ==",
            "blockIndex": "fb29934d-a688-4508-bfb6-29de21ac54d7"
        }
    ],
    "parentIndex": ""
}
```

##### Get all blocks in the Blockchain

```bash
$ curl http://localhost:8080/blockchain/blocks/all
```

Response:
```json
[
    {
        "index": "fb29934d-a688-4508-bfb6-29de21ac54d7",
        "timestamp": "2020-10-03T09:13:40.189489Z",
        "transactions": [
            {
                "index": "9aa55d17-8d4b-41da-8512-d89c1b8e8bc8",
                "sender": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
                "receiver": "-----BEGIN RSA PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA31qounIbnDNXw0Agdpfi\nFNBzaoR+QDsRV1JSy3euotRSDszYjEus93sfleScZNwx8IucceRJ77t0L7YeSp9d\nzRy69Y/zsX3k3X7czHkvM1CS/xx5nLbl77ie8Jn2GtSdPcVPeww4z9n7NB6ysvRQ\nS1aFQ97Gx3l7Wl3Kd6B/rywKVTmgjd+Nh6Kkl1+QMaaq6UhQKwqpcv07A+WUXmWI\nYgj/f5s2kao7XcC/6jBm8E7yj6OImAs4giWL4jufDrmrwtM6zfTCnGV7MfgR6qpD\no6e6xxBCsxYYIWMmxIFWjfU6i7C29S3zXes+p7VppvPLq3nuqWmkoamcrVYhXY6w\n5wIDAQAB\n-----END RSA PUBLIC KEY-----\n",
                "timestamp": "2020-10-03T09:13:37.26202Z",
                "data": "SW5jcm95YWJsZQ==",
                "blockIndex": "fb29934d-a688-4508-bfb6-29de21ac54d7"
            }
        ],
        "parentIndex": ""
    }
]
```

### Documentation

Launch a local http server with the code documentation using

```bash
$ godoc -http=:6060
```

Then navigate to [localhost:6060/pkg/github.com/peerbridge/peerbridge/](http://localhost:6060/pkg/github.com/peerbridge/peerbridge/)

### Deployment

#### Docker

As we are leveraging [BuildKit](https://github.com/moby/buildkit), you will need to make sure that you enable it by using Docker 19.03 or later and setting `DOCKER_BUILDKIT=1` in your environment. On Linux, macOS, or using WSL 2 you can do this using the following command:

```bash
$ export DOCKER_BUILDKIT=1
```

(You might also want to add this to your `.bashrc/.zshrc` file using `echo export DOCKER_BUILDKIT=1 > ~/.bashrc`)

On Windows for PowerShell you can use:
```powershell
$env:DOCKER_BUILDKIT=1
```

Or for command prompt:
```cmd
set DOCKER_BUILDKIT=1
```

We build two different stages for Unix-like OSes  (bin-unix) and for Windows (bin-windows).
We add aliases for Linux (bin-linux) and macOS (bin-darwin).
This allows us to make a dynamic target (bin) that depends on the `TARGETOS` variable and is automatically
set by the docker build platform flag.

Then simply build the application

Unix-like OSes (incl. Linux/MacOS):

```bash
$ make build
$ file bin/peerbridge
bin/peerbridge: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib/ld-, not stripped
```

Microsoft Windows:

```bash
$ make build-windows
$ file bin/peerbridge.exe
bin/peerbridge.exe: PE32+ executable (console) x86-64 (stripped to external PDB), for MS Windows
```
