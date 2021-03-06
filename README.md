# deps prints the dependencies of a Go package

**Install**: `go get github.com/grahamking/deps`

## Usage
```
USAGE: deps <package> [-d deep|layers -lib -stdlib -short]
"deps" prints the internal dependencies of a Go package.

-d deep|layers  Display more / different information
 deep: print the dependencies of the dependencies, recursively.
 layers: display the dependency layers

-lib  Include libraries.
 By default deps ignores anything starting with github.com, bitbucket.org, etc,
 because those are libraries and you only care about your app. Add this flag
 to prevent this ignoring.

-stdlib  Include Go built-in packages.
 By default deps ignores Go standard library packages. Add this flag
 to prevent this ignoring.

-short  Trim the package you are analyzing off the front of dependencies.
 e.g.: github.com/coreos/etcd/config -> config.

<package> is a path exactly like you would use in your code in "import".
That package and all it's dependencies must be on findable (GOPATH or stdlib).
```

## Examples

Basic usage just prints one level of dependencies:

    $ deps os/signal -stdlib

will output:
```
Dependencies of os/signal
 os
 sync
 syscall
```

The `-stdlib` flag says to include the Go standard packages, which are usually not displayed.

Adding the `-d deep` does this recursively.

    $ deps io -stdlib -d deep

will output:
```
Dependencies of io
io
| errors
| sync
| | sync/atomic
| | | unsafe
| | unsafe
```

### Layers

The `-d layers` option organises the dependencies by layers. The ones listed in higher rows depend on the ones in lower rows.

    $ deps github.com/hashicorp/serf -d layers

will display
```
Dependencies of github.com/hashicorp/serf
0: github.com/hashicorp/serf 2
1: github.com/hashicorp/serf/command 2, github.com/hashicorp/serf/command/agent 1
2: github.com/hashicorp/serf/client 1
3: github.com/hashicorp/serf/serf 0
```

The number listed after the package name is the number of dependencies of that package. 'serf/serf' does not depend on anything (expect possibly third-party or stdlib package - there's flags to show those). The 'serf/client' package only depends on 'serf/serf'. And so on upwards.

Here's 'serf' again, but with third-party packages:

    $ deps github.com/hashicorp/serf -d layers -lib

outputs
```
Dependencies of github.com/hashicorp/serf
0: github.com/hashicorp/serf 3
1: github.com/hashicorp/serf/command 4, github.com/hashicorp/serf/command/agent 8
2: github.com/armon/mdns 1, github.com/hashicorp/serf/client 3, github.com/mitchellh/cli 0, github.com/mitchellh/mapstructure 0
3: github.com/hashicorp/logutils 0, github.com/hashicorp/serf/serf 3, github.com/miekg/dns 0
4: github.com/hashicorp/memberlist 2
5: github.com/armon/go-metrics 0, github.com/ugorji/go/codec 0
```

Notice how 'github.com/mitchellh/cli' is on row number 2, even though it has no imports. Position does not say anything about outgoing imports, it tells you about incoming, which are in layer above. It's position on row 2 simply means that something on row 1 depends on it. Think of it as packages being as high up (less incoming dependencies) as they can.

Here's a bigger example, using `-short` to trim the name of internal packages:

    $ deps github.com/coreos/etcd -d layers -short

gives
```
Dependencies of github.com/coreos/etcd
0: github.com/coreos/etcd 7
1: config 5
2: discovery 3, pkg/strings 0, server 12, third_party/github.com/BurntSushi/toml 0
3: http 0, metrics 1, mod 4, pkg/http 1, server/v1 4, server/v2 5, store/v2 3
4: mod/dashboard 2, mod/leader/v2 2, mod/lock/v2 3, store 2, third_party/github.com/rcrowley/go-metrics 0
5: error 0, log 1, mod/dashboard/resources 0, third_party/github.com/coreos/go-etcd/etcd 0, third_party/github.com/coreos/raft 2, third_party/github.com/gorilla/mux 1
6: third_party/github.com/coreos/go-log/log 2, third_party/github.com/coreos/raft/protobuf 1, third_party/github.com/gorilla/context 0
7: third_party/bitbucket.org/kardianos/osext 0, third_party/code.google.com/p/gogoprotobuf/proto 0, third_party/github.com/coreos/go-systemd/journal 0
```

## Misc

Feedback, pull requests, etc are most welcome.

License is GPL, see the header of `deps.go`.
