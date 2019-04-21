package kvstructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/docker/libkv/store"
)

// Transcode takes an initialized interface and puts the data in a kv
func Transcode(s interface{}, prefix string, kv store.Store) error {
	transcoder, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix(prefix),
	)
	if err != nil {
		return err
	}

	return transcoder.Transcode(s)
}

// NewTranscoder returns a new transcoder for the given configuration.
// Once a transcoder has been returned, the same interface must be used
func NewTranscoder(opts ...TranscoderOpt) (Transcoder, error) {
	options := new(TranscoderOpts)

	t := new(transcoder)
	t.opts = options

	// configure transcoder
	configureTranscoder(t, opts...)

	return t, nil
}

// TranscoderWithPrefix ...
func TranscoderWithPrefix(prefix string) func(o *TranscoderOpts) {
	return func(o *TranscoderOpts) {
		o.Prefix = prefix
	}
}

// TranscoderWithKV ...
func TranscoderWithKV(kv store.Store) func(o *TranscoderOpts) {
	return func(o *TranscoderOpts) {
		o.KV = kv
	}
}

// Transcode is transcoding a given raw value interface to data in a kv store
func (t *transcoder) Transcode(s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr {
		return errors.New("kvstructure: interface must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return errors.New("kvstructure: interface must be addressable (a pointer)")
	}

	return t.transcode("", reflect.ValueOf(s).Elem())
}

// transcode is doing the heavy lifting in the background
func (t *transcoder) transcode(name string, val reflect.Value) error {
	var err error
	valKind := getKind(reflect.Indirect(val))
	switch valKind {
	case reflect.String:
		err = t.transcodeString(name, val)
	case reflect.Bool:
		err = t.transcodeBool(name, val)
	case reflect.Int:
		err = t.transcodeInt(name, val)
	// case reflect.Uint:
	// 	err = t.transcodeUint(name, val)
	// case reflect.Float32:
	// 	err = t.transcodeFloat(name, val)
	case reflect.Struct:
		err = t.transcodeStruct(name, val)
	// doesnt make sense
	// case reflect.Slice:
	// 	err = t.transcodeSlice(name, val)
	default:
		// we have to work on here for value to pointed to
		return fmt.Errorf("kvstructure: unsupported type %s", valKind)
	}

	// should be nil
	return err
}

// transdecodeString
func (t *transcoder) transcodeString(name string, val reflect.Value) error {
	return t.putKVPair(name, []byte(val.String()))
}

// transcodeBool
func (t *transcoder) transcodeBool(name string, val reflect.Value) error {
	return t.putKVPair(name, []byte(fmt.Sprint(val)))
}

// transcodeInt
func (t *transcoder) transcodeInt(name string, val reflect.Value) error {
	return t.putKVPair(name, []byte(fmt.Sprint(val)))
}

// transdecodeUint
func (t *transcoder) transcodeUint(name string, val reflect.Value) error {
	return nil
}

// transdecodeFloat
func (t *transcoder) transcodeFloat(name string, val reflect.Value) error {
	return nil
}

// transdecodeStruct
func (t *transcoder) transcodeStruct(name string, val reflect.Value) error {
	valInterface := reflect.Indirect(val)
	valType := valInterface.Type()

	var wg sync.WaitGroup
	wg.Add(valType.NumField())

	errors := make([]error, 0)

	// The slice will keep track of all structs we'll be transcoding.
	// There can be more structs, if we have embedded structs that are squashed.
	structs := make([]reflect.Value, 1, 5)
	structs[0] = val

	type field struct {
		field reflect.StructField
		val   reflect.Value
		json  bool
	}
	fields := []field{}

	for len(structs) > 0 {
		structVal := structs[0]
		structs = structs[1:]

		for i := 0; i < valType.NumField(); i++ {
			fieldType := valType.Field(i)
			isJSON := false

			tagParts := strings.Split(fieldType.Tag.Get(t.opts.TagName), ",")
			for _, tag := range tagParts[1:] {
				// test here for squashing
				if tag == "json" {
					isJSON = true
				}
			}

			fields = append(fields, field{fieldType, structVal.Elem().Field(i), isJSON})
		}
	}

	// evaluate all fields
	for _, f := range fields {
		field, val, isJSON := f.field, f.val, f.json
		kv := strings.ToLower(field.Name)

		tag := field.Tag.Get(t.opts.TagName)
		tag = strings.SplitN(tag, ",", 2)[0]
		if tag != "" {
			kv = tag
		}

		kv = strings.Join([]string{name, kv}, "/")

		if !val.CanAddr() {
			wg.Done()

			continue
		}

		// 	// we deal with
		if isJSON {
			// remove field from group
			wg.Done()

			if !val.CanAddr() {
				continue
			}

			b, err := json.Marshal(val.Interface())
			if err != nil {
				errors = append(errors, err)
				continue
			}

			// write to kv
			if err := t.putKVPair(kv, b); err != nil {
				errors = append(errors, err)
			}

			continue
		}

		go func() {
			defer wg.Done()
			if err := t.transcode(kv, val); err != nil {
				errors = append(errors, err)
			}
		}()

	}

	wg.Wait()

	return nil
}

// transdecodeBasic transdecode a basic type (bool, int, strinc, etc.)
// and eventually sets it to the retrieved value
func (t *transcoder) transdecodeBasic(val reflect.Value) error {
	return nil
}

// putKVPair
func (t *transcoder) putKVPair(key string, value []byte) error {
	return t.opts.KV.Put(trailingSlash(t.opts.Prefix)+key, value, nil)
}

// configureTranscoder
func configureTranscoder(t *transcoder, opts ...TranscoderOpt) error {
	for _, o := range opts {
		o(t.opts)
	}

	if t.opts.Metadata != nil {
		if t.opts.Metadata.Keys == nil {
			t.opts.Metadata.Keys = make([]string, 0)
		}

		if t.opts.Metadata.Unused == nil {
			t.opts.Metadata.Unused = make([]string, 0)
		}
	}

	if t.opts.TagName == "" {
		t.opts.TagName = "kvstructure"
	}

	return nil
}
