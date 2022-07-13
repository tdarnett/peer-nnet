# peer-nnet

Lightweight peer-to-peer application _to train a distributed neural network via federated learning `*`_.

The implementation is inspired by [BrainTorrent][].

`*` _in development_

[braintorrent]: https://arxiv.org/pdf/1905.06731.pdf

## to run

```bash
$ go run . -seed [some seed]
$ # open a second terminal
$ go run . -peer [peer address] -seed [some other seed]
```

## future ideas

- [Collect peer churn data](https://github.com/willscott/ipfs-counter/blob/willscott/churn/main.go) to inform remediation approaches

## helpful links

- [Flexible mocking for testing in Go](https://medium.com/safetycultureengineering/flexible-mocking-for-testing-in-go-f952869e34f5)
- [libp2p in go](https://ldej.nl/post/building-an-echo-application-with-libp2p/)
