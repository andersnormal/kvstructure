package kvstructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/libkv/store"
)

// Transdecode takes an interface and uses reflection
// to fill it with data from a kv.
func Transdecode(s interface{}, prefix string, kv store.Store) error {
	transdecoder, err := NewTransdecoder(
		TransdecoderWithPrefix(prefix),
		TransdecoderWithKV(kv),
	)
	if err != nil {
		return err
	}

	return transdecoder.Transdecode(s)
}

// NewTransdecoder returns a new transdecoder for the given configuration.
// Once a transdecoder has been returned, the same configuration must not be used
// again.
func NewTransdecoder(opts ...TransdecoderOpt) (Transdecoder, error) {
	options := new(TransdecoderOpts)

	t := new(transdecoder)
	t.opts = options

	// configure transcoder
	configureTransdecoder(t, opts...)

	return t, nil
}

// TransdecoderWithPrefix ...
func TransdecoderWithPrefix(prefix string) func(o *TransdecoderOpts) {
	return func(o *TransdecoderOpts) {
		o.Prefix = prefix
	}
}

// TransdecoderWithKV ...
func TransdecoderWithKV(kv store.Store) func(o *TransdecoderOpts) {
	return func(o *TransdecoderOpts) {
		o.KV = kv
	}
}

// Transdecode transdecodes a given raw interface to a filled structure
func (t *transdecoder) Transdecode(s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr {
		return errors.New("kvstructure: interface must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return errors.New("kvstructure: interface must be addressable (a pointer)")
	}

	return t.transdecode("", reflect.ValueOf(s).Elem())
}

// transdecode is doing the heavy lifting in the background
func (t *transdecoder) transdecode(name string, val reflect.Value) error {
	var err error
	valKind := getKind(val)
	switch valKind {
	case reflect.String:
		err = t.transdecodeString(name, val)
	case reflect.Bool:
		err = t.transdecodeBool(name, val)
	case reflect.Int:
		err = t.transdecodeInt(name, val)
	case reflect.Uint:
		err = t.transdecodeUint(name, val)
	case reflect.Float32:
		err = t.transdecodeFloat(name, val)
	case reflect.Struct:
		err = t.transdecodeStruct(val)
	// case reflect.Slice:
	// 	err = t.transdecodeSlice(name, val)
	default:
		// we have to work on here for value to pointed to
		return fmt.Errorf("kvstructure: unsupported type %s", valKind)
	}

	// should be nil
	return err
}

// transdecodeBasic transdecode a basic type (bool, int, strinc, etc.)
// and eventually sets it to the retrieved value
func (t *transdecoder) transdecodeBasic(val reflect.Value) error {
	return nil
}

// transdecodeString
func (t *transdecoder) transdecodeString(name string, val reflect.Value) error {
	kvPair, err := t.getKVPair(name)
	if err != nil {
		return err
	}
	kvVal := string(kvPair.Value)

	conv := true
	switch {
	case val.Kind() == reflect.String:
		val.SetString(kvVal)
	default:
		conv = false
	}

	// if conf was not successful
	if !conv {
		return err
	}

	return nil
}

// transdecodeBool
func (t *transdecoder) transdecodeBool(name string, val reflect.Value) error {
	kvPair, err := t.getKVPair(name)
	if err != nil {
		return err
	}
	kvVal := string(kvPair.Value)

	switch {
	case val.Kind() == reflect.Bool:
		conv, err := strconv.ParseBool(kvVal)
		if err != nil {
			return err
		}
		val.SetBool(conv)
	default:
		return fmt.Errorf("'%s' got unconvertible type '%s'", name, val.Type())
	}

	return nil
}

// transdecodeInt
func (t *transdecoder) transdecodeInt(name string, val reflect.Value) error {
	kvPair, err := t.getKVPair(name)
	if err != nil {
		return err
	}
	kvVal := string(kvPair.Value)

	switch {
	case val.Kind() == reflect.Int:
		conv, err := strconv.ParseInt(kvVal, 10, 64)
		if err != nil {
			return err
		}
		val.SetInt(conv)
	case val.Kind() == reflect.Uint:
		conv, err := strconv.ParseUint(kvVal, 10, 64)
		if err != nil {
			return err
		}
		val.SetUint(conv)
	default:
		return fmt.Errorf("'%s' got unconvertible type '%s'", name, val.Type())
	}

	return nil
}

// transdecodeUint
func (t *transdecoder) transdecodeUint(name string, val reflect.Value) error {
	kvPair, err := t.getKVPair(name)
	if err != nil {
		return err
	}
	kvVal := string(kvPair.Value)

	switch {
	case val.Kind() == reflect.Uint:
		conv, err := strconv.ParseUint(kvVal, 10, 64)
		if err != nil {
			return err
		}
		val.SetUint(conv)
	default:
		return fmt.Errorf("'%s' got unconvertible type '%s'", name, val.Type())
	}

	return nil
}

// transdecodeFloat32
func (t *transdecoder) transdecodeFloat(name string, val reflect.Value) error {
	kvPair, err := t.getKVPair(name)
	if err != nil {
		return err
	}
	kvVal := string(kvPair.Value)

	switch {
	case val.Kind() == reflect.Float32:
		conv, err := strconv.ParseFloat(kvVal, 64)
		if err != nil {
			return err
		}
		val.SetFloat(conv)
	default:
		return fmt.Errorf("'%s' got unconvertible type '%s'", name, val.Type())
	}

	return nil
}

// transdecodeStruct
func (t *transdecoder) transdecodeStruct(val reflect.Value) error {
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
	for len(structs) > 0 { // could be easier
		structVal := structs[0]
		structs = structs[1:]
		// here we should do squashing

		for i := 0; i < valType.NumField(); i++ {
			fieldType := valType.Field(i)
			// json is somehow special
			// it is curated by golang json
			isJSON := false
			// fieldKind := fieldType.Type.Kind()

			tagParts := strings.Split(fieldType.Tag.Get(t.opts.TagName), ",")
			for _, tag := range tagParts[1:] {
				// test here for squashing
				if tag == "json" {
					isJSON = true
				}
			}

			fields = append(fields, field{fieldType, structVal.Field(i), isJSON})
		}
	}

	for _, f := range fields {
		field, val, isJSON := f.field, f.val, f.json
		kv := field.Name

		tag := field.Tag.Get(t.opts.TagName)
		tag = strings.SplitN(tag, ",", 2)[0]
		if tag != "" {
			kv = tag
		}

		if !val.CanSet() {
			wg.Done()

			continue
		}

		// we deal with
		if isJSON {
			// remove field from group
			wg.Done()

			if !val.CanAddr() {
				continue
			}

			kvPair, err := t.getKVPair(kv)
			if err != nil {
				errors = append(errors, err)
				continue
			}

			obj := reflect.New(field.Type).Interface()
			if err := json.Unmarshal(kvPair.Value, &obj); err != nil {
				errors = append(errors, err)
			}

			val.Set(reflect.ValueOf(obj).Elem())

			continue
		}

		go func() {
			defer wg.Done()
			if err := t.transdecode(kv, val); err != nil {
				errors = append(errors, err)
			}
		}()
	}

	wg.Wait()

	return nil
}

func (t *transdecoder) getKVPair(key string) (*store.KVPair, error) {
	kvPair, err := t.opts.KV.Get(trailingSlash(t.opts.Prefix) + key)
	if err != nil {
		return nil, err
	}

	return kvPair, nil
}

// getKind is returning the kind of the reflected value
func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}

// configureTransdecoder
func configureTransdecoder(t *transdecoder, opts ...TransdecoderOpt) error {
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
