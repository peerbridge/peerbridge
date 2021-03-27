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

This will automatically generate a new random ECDSA secp256k1 key pair and save it under `key.json`. Do not share this key pair - it is your authentication to the blockchain network!

Note that this will start your node independently from other nodes! If you want to connect to an existing bootstrap node within the blockchain network (or another node that you started locally), use the `-r` option to specify a remote node, as follows:

```bash
$ REMOTE_URL="http://peerbridge.herokuapp.com" go run main.go
```

You can also specify a custom key path:

```bash
$ KEY_PATH="./my.key.json" go run main.go
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
