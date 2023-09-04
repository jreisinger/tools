Tools are small Go programs for various sysadmin-style tasks done from terminal.

To use a tool on a local machine:

```
$ go doc cmd/<tool>
$ go build -o ~/bin/ cmd/<tool>/*
$ <tool>
```

To use a tool on a remote machine:

```
$ go tool dist list
$ GOOS=linux GOARCH=arm64 go build cmd/<tool>/*
$ scp ./<tool> user@raspberry.net:
$ ssh user@raspberry.net ./<tool>
```
