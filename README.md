## inspired by

- https://ldej.nl/post/building-an-echo-application-with-libp2p/
- https://arxiv.org/pdf/1905.06731.pdf

## to run

```bash
$ go run . - seed [some seed]
$ # open a second terminal
$ go run . -peer [peer address] - seed [some other seed]
```

## future ideas

- [Collect peer churn data](https://github.com/willscott/ipfs-counter/blob/willscott/churn/main.go) to inform remediation approaches

## helpful links

- [Flexible mocking for testing in Go](https://medium.com/safetycultureengineering/flexible-mocking-for-testing-in-go-f952869e34f5)
