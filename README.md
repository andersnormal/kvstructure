# kvstructure

[![GoDoc](https://godoc.org/github.com/andersnormal/kvstructure?status.png)](https://godoc.org/github.com/andersnormal/kvstructure)
[![Build Status](https://travis-ci.org/andersnormal/kvstructure.svg?branch=master)](https://travis-ci.org/andersnormal/kvstructure)
[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)
[![Go Report Card](https://goreportcard.com/badge/github.com/andersnormal/kvstructure)](https://goreportcard.com/report/github.com/andersnormal/kvstructure)

Go library for transcoding data from KVs supported by [libkv](https://github.com/docker/libkv) to `structs`, `string`, `int`, `uint` and `float32` and vice versa.

## Example

```golang
transcoder, err := NewTranscoder(
	TranscoderWithKV(kv),
	TranscoderWithPrefix("prefix"),
)
    
if err != nil {
    return err
}

tt := &Example{
	Description: "bar",
	Enabled: true,
}

if err := transcoder.Transcode("foo", &tt) {
    return err
}
```

## License
[Apache 2.0](/LICENSE)
