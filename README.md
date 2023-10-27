Tools are small Go programs and functions that I find useful or entertaining.

To use a tool on a local machine:

```
$ go install ./cmd/<tool>
$ <tool>
```

To use a tool on a remote machine:

```
$ go tool dist list
$ GOOS=linux GOARCH=arm64 go build -o /tmp/ ./cmd/<tool>
$ scp /tmp/<tool> user@raspberry.net:
$ ssh user@raspberry.net ./<tool>
```

To use a function do a copy/paste since I don't care about backward compatibility here.