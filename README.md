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

Then, start the PeerBridge Blockchain Server:
```bash
$ go run main.go
```

This will automatically generate a new random ECDSA secp256k1 key pair and save it under `./key.json`. Do not share this key pair! It is your authentication to the blockchain network!

Note that this will start your node independently of other nodes! If you want to connect to an existing bootstrap node within the blockchain network (or another node that you started locally), use the `-r` option to specify a remote node, as follows:

```bash
$ REMOTE_URL="http://peerbridge.herokuapp.com" go run main.go
```

You can also specify a custom key path:

```bash
$ KEY_PATH="./my.key.json" go run main.go
```

You can also (alternatively) give your private key in the environment:

```bash
$ PRIVATE_KEY="484ff6fe0382d9f0c201d3f7a7e65e2a4f86845ccc47bc5b8617b31666ddf408" go run main.go
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

We use a multistage approach together with the docker build platform flag to build binaries for the different operating systems.
Currently, Linux and Windows binaries for amd64 (64-Bit) architecture can be cross compiled.

Then simply build the application using make:

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
