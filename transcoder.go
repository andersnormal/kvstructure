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
	config := &TranscoderConfig{
		Prefix:   prefix,
		KV:       kv,
		Metadata: nil,
		Input:    s,
	}

	transcoder, err := NewTranscoder(config)
	if err != nil {
		return err
	}

	return transcoder.Transcode()
}

// NewTranscoder returns a new transcoder for the given configuration.
// Once a transcoder has been returned, the same interface must be used
func NewTranscoder(config *TranscoderConfig) (*Transcoder, error) {
	val := reflect.ValueOf(config.Input)
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("input muse be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return nil, errors.New("input must be addressable (a pointer)")
	}

	if config.Metadata != nil {
		if config.Metadata.Keys == nil {
			config.Metadata.Keys = make([]string, 0)
		}

		if config.Metadata.Unused == nil {
			config.Metadata.Unused = make([]string, 0)
		}
	}

	if config.TagName == "" {
		config.TagName = "kvstructure"
	}

	result := &Transcoder{
		config: config,
	}

	return result, nil
}

// Transcode is transcoding a given raw value interface to data in a kv store
func (t *Transcoder) Transcode() error {
	return t.transcode("", reflect.ValueOf(t.config.Input).Elem())
}

// transcode is doing the heavy lifting in the background
func (t *Transcoder) transcode(name string, val reflect.Value) error {
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
		return fmt.Errorf("Unsupported type %s", valKind)
	}

	// should be nil
	return err
}

// transdecodeString
func (t *Transcoder) transcodeString(name string, val reflect.Value) error {
	return t.putKVPair(name, []byte(val.String()))
}

// transcodeBool
func (t *Transcoder) transcodeBool(name string, val reflect.Value) error {
	return t.putKVPair(name, []byte(fmt.Sprint(val)))
}

// transcodeInt
func (t *Transcoder) transcodeInt(name string, val reflect.Value) error {
	return t.putKVPair(name, []byte(fmt.Sprint(val)))
}

// transdecodeUint
func (t *Transcoder) transcodeUint(name string, val reflect.Value) error {
	return nil
}

// transdecodeFloat
func (t *Transcoder) transcodeFloat(name string, val reflect.Value) error {
	return nil
}

// transdecodeStruct
func (t *Transcoder) transcodeStruct(name string, val reflect.Value) error {
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

			tagParts := strings.Split(fieldType.Tag.Get(t.config.TagName), ",")
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

		tag := field.Tag.Get(t.config.TagName)
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
func (t *Transcoder) transdecodeBasic(val reflect.Value) error {
	return nil
}

// putKVPair
func (t *Transcoder) putKVPair(key string, value []byte) error {
	return t.config.KV.Put(trailingSlash(t.config.Prefix)+key, value, nil)
}
