[https://peerbridge.herokuapp.com](https://peerbridge.herokuapp.com/)

**NOTICE**

UNDER DEVELOPMENT

##  Installation

Install the PeerBridge CLI with the `go get` command:

```bash
$ go get -u github.com/peerbridge/peerbridge
```

## Quick Start

To bootstrap a PeerBridge Blockchain Server you need a ECDSA secp256k1 key pair. 
You can create a new secp256k1 key pair using: 

```bash
$ peerbridge key create -s

Successfully created a new secp256k1 keypair.

Private Key: 52554fb4acafd830b89698f995e52edf52b60136df86d86045a54ede32478838
Public Key: 0253ec55286c03273061612a064e7d85feab4f26e7a2cf5cde3f90cb368b5c05c6

Successfully saved config file: /home/felix/.peerbridge.yml
```

This will create a new config file located in your home directory.
Do not share this key pair! It is your authentication to the blockchain network!

The PeerBridge Blockchain Server requires a PostgreSQL Database.
Start a new PostgreSQL Database instance locally using Docker.

```bash
$ docker run --name postgres -p 5432:5432 \
    -e POSTGRES_DB=peerbridge \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=password \
    -d postgres:13-alpine
```

Now you can start the PeerBridge Blockchain Server. 

```bash
$ export DATABASE_URL=postgres://postgres:password@localhost:5432/postgres?sslmode=disable 
$ peerbridge server --host https://peerbridge.herokuapp.com --sync
```

This will connect your Blockchain Server to the peer node [https://peerbridge.herokuapp.com](https://peerbridge.herokuapp.com) 
and sync all blocks.


## Deployment

Start the PeerBridge Blockchain Server including PostgreSQL Database and Adminer using the docker-compose deployment
provided with the source code of this repository.

```bash
$ git clone https://github.com/peerbridge/peerbridge.git
$ cd peerbridge
$ docker-compose up -d
```

### Development

Checkout the sources from GitHub.

```bash
$ git clone https://github.com/peerbridge/peerbridge.git
$ cd peerbridge
```

Start the PostgreSQL Database and Adminer using
```bash
$ docker-compose up -d database adminer
```

PeerBridge consists of a CLI application build with [spf13/cobra](https://github.com/spf13/cobra) and [spf13/viper](https://github.com/spf13/viper). To invoke it use:

```bash
$ go run main.go 

PeerBridge is a proof-of-stake based blockchain cryptosystem,
as a foundation for the experimental PeerBridge messenger.

Usage:
  peerbridge [command]

Available Commands:
  help        Help about any command
  key         Manage secp256k1 keys
  node        View details about PeerBridge nodes
  server      Start a new blockchain node
  transaction Manage transactions inside the blockchain

Flags:
      --config string   config file (default is $HOME/.peerbridge.yaml)
  -h, --help            help for peerbridge

Use "peerbridge [command] --help" for more information about a command.
```

To bootstrap a PeerBridge Blockchain Server you need a ECDSA secp256k1 key pair. 
You can create a new secp256k1 key pair using: 

```bash
$ go run main.go key create -s

Successfully created a new secp256k1 keypair.

Private Key: 52554fb4acafd830b89698f995e52edf52b60136df86d86045a54ede32478838
Public Key: 0253ec55286c03273061612a064e7d85feab4f26e7a2cf5cde3f90cb368b5c05c6

Successfully saved config file: ~/.peerbridge.yml
```

This will create a new config file located in your home directory under `~/.peerbridge.yml`.
Afterwards, if you invoke the `peerbridge` CLI the config file will be automatically picked 
up by [spf13/viper](https://github.com/spf13/viper). Thus, the generated key will be used
for subsequent operations made with the CLI. See [Configuration File](#configuration-file).

Now you can start a new PeerBridge Blockchain Server. 

```bash
$ export DATABASE_URL=postgres://postgres:password@localhost:5432/postgres?sslmode=disable 
$ go run main.go server
```

If you want to sync your local PeerBridge Blockchain Server against a remote node, simply specify the `--host` option
and set the `--sync` flag.

Example:
```bash
$ export DATABASE_URL=postgres://postgres:password@localhost:5432/postgres?sslmode=disable 
$ go run main.go server --host https://peerbridge.herokuapp.com --sync
```

This will connect your Blockchain Server to the peer node [https://peerbridge.herokuapp.com](https://peerbridge.herokuapp.com) 
and sync all blocks.

## CLI

### Usage

```bash
$ go run main.go --help

PeerBridge is a proof-of-stake based blockchain cryptosystem,
as a foundation for the experimental PeerBridge messenger.

Usage:
  peerbridge [command]

Available Commands:
  help        Help about any command
  key         Manage secp256k1 keys
  node        View details about PeerBridge nodes
  server      Start a new blockchain node
  transaction Manage transactions inside the blockchain

Flags:
      --config string   config file (default is $HOME/.peerbridge.yaml)
  -h, --help            help for peerbridge

Use "peerbridge [command] --help" for more information about a command.
```

### Key

Create a new secp256k1 key pair.

```bash
$ go run main.go key create --help
Create a new secp256k1 keypair for usage inside the PeerBridge blockchain.

Usage:
  peerbridge key create [flags]

Flags:
  -h, --help   help for create
  -s, --save   save generated key to config file (default is $HOME/.peerbridge.yaml)

Global Flags:
      --config string   config file (default is $HOME/.peerbridge.yaml)
```

```bash
$ go run main.go key create -s
Using config file: /home/felix/.peerbridge.yml

Successfully created a new secp256k1 keypair.

Private Key: 64c42cf769cf05c4b26a3835403e66cec1a4f2ff835475834014f7f0dde415ed
Public Key: 02eddd74ed8637e32621c72a36485157b17f51858f5051cdbe056f005c8389f5c4

Successfully saved config file: /home/felix/.peerbridge.yml 
```

### Balance

Retrieve the current account balance

```bash
$ go run main.go node balance --help
Retrieve the account balance of a node inside the PeerBridge blockchain.

Usage:
  peerbridge node balance [flags]

Flags:
  -h, --help   help for balance

Global Flags:
      --config string   config file (default is $HOME/.peerbridge.yaml)
      --host string     blockchain node to connect to (default "https://peerbridge.herokuapp.com")
      --key string      secp256k1 key of the account
```

Example:

```bash
$ go run main.go node balance --host https://peerbridge.herokuapp.com --key 0372689db204d56d9bb7122497eef4732cce308b73f3923fc076aed3c2dfa4ad04
Checking account balance for key 0372689db204d56d9bb7122497eef4732cce308b73f3923fc076aed3c2dfa4ad04 on host https://peerbridge.herokuapp.com.
Account balance: 101000
```

```bash
$ go run main.go node balance --host https://peerbridge.herokuapp.com --key eba4f82788edb8e464920293ff06605484bef87561880e44b6e4902f27e6d6ca
Generating public key.
Checking account balance for key 0372689db204d56d9bb7122497eef4732cce308b73f3923fc076aed3c2dfa4ad04 on host https://peerbridge.herokuapp.com.
Your account balance: 101000
```

### Transaction

Create a new transaction.

```bash
$ go run main.go transaction create --help
Create new transactions and submit them to the PeerBridge blockchain.

Usage:
  peerbridge transaction create [flags]

Flags:
      --amount uint       Amount to transfer as part of the transaction
  -h, --help              help for create
      --receiver string   secp256k1 public key of the receiver of the transaction
      --sender string     secp256k1 private key of the account to create a transaction

Global Flags:
      --config string   config file (default is $HOME/.peerbridge.yaml)
      --host string     blockchain node to connect to (default "https://peerbridge.herokuapp.com")
```

```bash
$ go run main.go transaction create --host https://peerbridge.herokuapp.com --sender eba4f82788edb8e464920293ff06605484bef87561880e44b6e4902f27e6d6ca --receiver 03f1f2fbd80b49b8ffc8194ac0a0e0b7cf0c7e21bca2482c5fba7adf67db41dec5 --amount 420 
```

### Server

```bash
$ go run main.go server --help
Start a new PeerBridge blockchain node on the current host

Usage:
  peerbridge server [flags]

Flags:
  -h, --help          help for server
      --host string   blockchain node to connect to (default "https://peerbridge.herokuapp.com")
      --key string    secp256k1 key of the account
      --sync          sync the server against the specified host (default is https://peerbridge.herokuapp.com) (default true)

Global Flags:
      --config string   config file (default is $HOME/.peerbridge.yaml)
```

Example:

```bash
$ go run main.go server --key eba4f82788edb8e464920293ff06605484bef87561880e44b6e4902f27e6d6ca
```
```bash
$ go run main.go server --key eba4f82788edb8e464920293ff06605484bef87561880e44b6e4902f27e6d6ca --host https://peerbridge.herokuapp.com  --sync
```

## Configuration File

Both `key` and `host` command line flags provided to the various commands can be provided using a
config file instead of manually passing them to each command. Per default PeerBridge will try to 
locate a `.peerbridge.yaml` file located in your home directory. 

Example:

`~/.peerbridge.yaml`

```yaml
host: https://peerbridge.herokuapp.com
key: eba4f82788edb8e464920293ff06605484bef87561880e44b6e4902f27e6d6ca
```

Otherwise, you can specify a custom config file located in another directory using the `--config` flag.

Example:

```bash
$ peerbridge node balance --config /my/config/path/.peerbridge.yaml
```

Instead of manually creating a new config, use the [key command](#key) with the `--save` flag.
For more options on how to specify these flags using e.g. environment variables or using a 
key/value store refer to the [viper documentation](https://github.com/spf13/viper).

## Documentation

Launch a local http server with the code documentation using

```bash
$ godoc -http=:6060
```

Then navigate to [localhost:6060/pkg/github.com/peerbridge/peerbridge/](http://localhost:6060/pkg/github.com/peerbridge/peerbridge/)

## Build

### Docker

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

We use a multistage approach together with the docker build platform flag to build binaries for the different operating systems.
Currently, Linux and Windows binaries for amd64 (64-Bit) architecture can be cross compiled.

Then simply build the application using make:

Linux:

```bash
$ make bin-linux
$ file bin/peerbridge-linux-amd64
bin/peerbridge-linux-amd64: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=..., stripped
```

Microsoft Windows:

```bash
$ make bin-windows
$ file bin/peerbridge-windows-amd64.exe
bin/peerbridge-windows-amd64.exe: PE32+ executable (console) x86-64 (stripped to external PDB), for MS Windows
```
