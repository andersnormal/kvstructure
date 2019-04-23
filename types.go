package kvstructure

import "github.com/docker/libkv/store"

// Transcoder is the interface to a transcoder
type Transcoder interface {
	Transcode(string, interface{}) error
}

// Transdecoder is the interface to a transdecoder
type Transdecoder interface {
	Transdecode(string, interface{}) error
}

// TransdecoderOpt ...
type TransdecoderOpt func(*TransdecoderOpts)

// TransdecoderOpts is the configuration that is used to create a new transdecoder
// and allows customization of various aspects of decoding.
type TransdecoderOpts struct {
	// ZeroFields, if set to true, will zero fields before writing them.
	// For example, a map will be emptied before decoded values are put in
	// it. If this is false, a map will be merged.
	ZeroFields bool

	// should be done later
	WeaklyTypedInput bool

	// Metadata is the struct that will contain extra metadata about
	// the decoding. If this is nil, then no metadata will be tracked.
	Metadata *Metadata

	// The tag name that kvstructure reads for field names. This
	// defaults to "kvstructure"
	TagName string

	// Prefix is the prefix of the store
	Prefix string

	// KV is the kv used to retrieve the needed infos
	KV store.Store
}

// TranscoderOpt ...
type TranscoderOpt func(*TranscoderOpts)

// TranscoderOpts is the configuration that is used to create a new transcoder
// and allows customization of various aspects of decoding.
type TranscoderOpts struct {
	// Metadata is the struct that will contain extra metadata about
	// the decoding. If this is nil, then no metadata will be tracked.
	Metadata *Metadata

	// The tag name that kvstructure reads for field names. This
	// defaults to "kvstructure"
	TagName string

	// Prefix is the prefix of the store
	Prefix string

	// KV is the kv used to retrieve the needed infos
	KV store.Store
}

// A Transdecoder takes a raw interface value and turns it into structured data
type transdecoder struct {
	opts *TransdecoderOpts
}

// A Transcoder takes a raw interface and puts it into a kv structure
type transcoder struct {
	opts *TranscoderOpts
}

// Metadata contains information about decoding a structure that
// is tedious or difficult to get otherwise.
type Metadata struct {
	// Keys are the keys of the structure which were successfully decoded
	Keys []string

	// Unused is a slice of keys that were found in the raw value but
	// weren't decoded since there was no matching field in the result interface
	Unused []string
}
