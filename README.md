Tools are small Go programs for various sysadmin-style tasks done from terminal.

To use a tool:

```
$ go doc cmd/<tool>/<tool>.go
$ go run cmd/<tool>/<tool>.go
```

or 

```
$ go install cmd/<tool>/<tool>.go # installs into ~/go/bin by default
$ <tool>
```

or

```
$ GOOS=linux GOARCH=arm64 go build cmd/<tool>/<tool>.go # go tool dist list
$ scp ./<tool> user@raspberry.net:
$ ssh user@raspberry.net
raspberry$ ./<tool>
```