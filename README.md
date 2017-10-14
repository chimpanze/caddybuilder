# CaddyBuilder
Command line utility for building [Caddy Web Server](https://github.com/mholt/caddy).

You will need git and go installed on your machine.

### License

This tool will build Caddy from [source code](https://github.com/mholt/caddy). As such, the executable you get by using this tool is subject to the project's [Apache 2.0 license](https://github.com/mholt/caddy/blob/master/LICENSE.txt), but it neither contains nor is subject to [the EULA for Caddy's official binary distributions](https://github.com/mholt/caddy/blob/master/dist/EULA.txt).

## Usage
```
  -dev
        Build the current master branch
  -goarch string
        ARCH for which to build
  -goos string
        OS for which to build
  -plugin value
        Plugin to integrate in the build
```

## Example
```
go run caddybuilder.go -goos linux -goarch amd64 -plugin expires -plugin filemanager
```

## Useful info
You can find the list of compatible plugins [here](https://github.com/mholt/caddy/blob/baf6db5b570e36ea2fee30d50f879255a5895370/caddyhttp/httpserver/plugin.go#L448).

List of GOOS and GOARCH possible values [here](https://github.com/golang/go/blob/master/src/go/build/syslist.go).
