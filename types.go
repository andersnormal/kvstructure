package kvstructure

import "github.com/docker/libkv/store"

// TransdecoderConfig is the configuration that is used to create a new transdecoder
// and allows customization of various aspects of decoding.
type TransdecoderConfig struct {
	// mutex *sync.RWMutex

	// ZeroFields, if set to true, will zero fields before writing them.
	// For example, a map will be emptied before decoded values are put in
	// it. If this is false, a map will be merged.
	ZeroFields bool

	// should be done later
	WeaklyTypedInput bool

	// Metadata is the struct that will contain extra metadata about
	// the decoding. If this is nil, then no metadata will be tracked.
	Metadata *Metadata

	// Result is a pointer to the struct that will contain the decoded
	// value.
	Result interface{}

	// The tag name that kvstructure reads for field names. This
	// defaults to "kvstructure"
	TagName string

	// Prefix is the prefix of the store
	Prefix string

	// KV is the kv used to retrieve the needed infos
	KV store.Store
}

// TranscoderConfig is the configuration that is used to create a new transcoder
// and allows customization of various aspects of decoding.
type TranscoderConfig struct {
	// mutex *sync.RWMutex

	// Result is a pointer to the struct that will contain the decoded
	// value.
	Input interface{}

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
type Transdecoder struct {
	config *TransdecoderConfig
}

// A Transcoder takes a raw interface and puts it into a kv structure
type Transcoder struct {
	config *TranscoderConfig
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
