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

