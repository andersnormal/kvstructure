package kvstructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/docker/libkv/store"
	"golang.org/x/sync/errgroup"
)

// Transdecode takes an interface and uses reflection
// to fill it with data from a kv.
func Transdecode(name string, s interface{}, prefix string, kv store.Store) error {
	transdecoder, err := NewTransdecoder(
		TransdecoderWithPrefix(prefix),
		TransdecoderWithKV(kv),
	)
	if err != nil {
		return err
	}

	return transdecoder.Transdecode(name, s)
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
func (t *transdecoder) Transdecode(name string, s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr {
		return errors.New("kvstructure: interface must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return errors.New("kvstructure: interface must be addressable (a pointer)")
	}

	return t.transdecode(name, reflect.ValueOf(s).Elem(), nil)
}

// transdecode is doing the heavy lifting in the background
func (t *transdecoder) transdecode(name string, val reflect.Value, kvPair *store.KVPair) error {
	var err error
	valKind := getKind(reflect.Indirect(val))
	switch valKind {
	case reflect.String:
		err = t.transdecodeString(name, val, kvPair)
	case reflect.Bool:
		err = t.transdecodeBool(name, val, kvPair)
	case reflect.Int:
		err = t.transdecodeInt(name, val, kvPair)
	case reflect.Uint:
		err = t.transdecodeUint(name, val, kvPair)
	case reflect.Float32:
		err = t.transdecodeFloat(name, val, kvPair)
	case reflect.Struct:
		err = t.transdecodeStruct(name, val)
	case reflect.Slice:
		// silent do nothing
		err = t.transdecodeSlice(name, val)
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
func (t *transdecoder) transdecodeString(name string, val reflect.Value, kvPair *store.KVPair) error {
	kvPair, err := t.getKVPair(name, kvPair)
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
func (t *transdecoder) transdecodeBool(name string, val reflect.Value, kvPair *store.KVPair) error {
	kvPair, err := t.getKVPair(name, kvPair)
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
func (t *transdecoder) transdecodeInt(name string, val reflect.Value, kvPair *store.KVPair) error {
	kvPair, err := t.getKVPair(name, kvPair)
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
func (t *transdecoder) transdecodeUint(name string, val reflect.Value, kvPair *store.KVPair) error {
	kvPair, err := t.getKVPair(name, kvPair)
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
func (t *transdecoder) transdecodeFloat(name string, val reflect.Value, kvPair *store.KVPair) error {
	kvPair, err := t.getKVPair(name, kvPair)
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
func (t *transdecoder) transdecodeStruct(name string, val reflect.Value) error {
	valInterface := reflect.Indirect(val)
	valType := valInterface.Type()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create an errgroup to trace the latest error and return
	g, _ := errgroup.WithContext(ctx)

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

			// detected if this field is json
			if fieldType.Tag.Get("json") != "" {
				isJSON = true
			}

			fields = append(fields, field{fieldType, structVal.Field(i), isJSON})
		}
	}

	for _, f := range fields {
		f := f
		field, val, isJSON := f.field, f.val, f.json
		kv := strings.ToLower(field.Name)

		tag := field.Tag.Get(t.opts.TagName)
		tag = strings.SplitN(tag, ",", 2)[0]
		if tag != "" {
			kv = tag
		}

		if name != "" {
			kv = strings.Join([]string{name, kv}, "/")
		}

		if !val.CanSet() {
			continue
		}

		// we deal with
		if isJSON && tag == "" {
			if !val.CanAddr() {
				continue
			}

			// check if we have to omit
			tag := field.Tag.Get("json")
			if tag == "-" {
				continue
			}

			g.Go(func() error {
				// if there is no kvPair
				kvPair, err := t.getKVPair(kv, nil)
				if err != nil {
					return err
				}

				obj := reflect.New(field.Type).Interface()
				if err := json.Unmarshal(kvPair.Value, &obj); err != nil {
					return fmt.Errorf("'%s' field got : %s", field.Name, err)
				}

				if obj == nil {
					return nil
				}

				val.Set(reflect.ValueOf(obj).Elem())

				return nil
			})

			continue
		}

		g.Go(func() error {
			if err := t.transdecode(kv, val, nil); err != nil {
				return err
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

// transdecodeSlice
func (t *transdecoder) transdecodeSlice(name string, val reflect.Value) error {
	kvPairs, err := t.listKVPairs(name)
	if err != nil {
		return err
	}

	s := reflect.MakeSlice(val.Type(), len(kvPairs), len(kvPairs))
	val.Set(s)

	// todo: this can be more efficient, because this is costly
	for i, v := range kvPairs {
		kind := getKind(val.Index(i))
		switch kind {
		case reflect.Ptr:
			val.Index(i).Set(reflect.New(val.Index(i).Type().Elem()))
			t.transdecode(strings.Replace(v.Key, trailingSlash(t.opts.Prefix), "", -1), val.Index(i).Elem(), nil)
		case reflect.String:
			fallthrough
		case reflect.Bool:
			fallthrough
		case reflect.Int:
			fallthrough
		case reflect.Uint:
			fallthrough
		case reflect.Float32:
			fallthrough
		case reflect.Slice:
			t.transdecode(strings.Replace(v.Key, trailingSlash(t.opts.Prefix), "", -1), val.Index(i), v)
		default:
			return fmt.Errorf("'%s' got unconvertible type '%s'", name, val.Type())
		}
	}

	return nil
}

func like(like interface{}) interface{} {
	typ := reflect.TypeOf(like)
	one := reflect.New(typ)

	return one.Interface()
}

// getKVPair
func (t *transdecoder) getKVPair(key string, kvPair *store.KVPair) (*store.KVPair, error) {
	var err error

	if kvPair != nil {
		return kvPair, nil
	}

	kvPair, err = t.opts.KV.Get(trailingSlash(t.opts.Prefix) + key)
	if err != nil {
		return nil, err
	}

	return kvPair, nil
}

func (t *transdecoder) listKVPairs(key string) ([]*store.KVPair, error) {
	kvPairs, err := t.opts.KV.List(trailingSlash(t.opts.Prefix) + key)
	if err != nil {
		return nil, err
	}

	return kvPairs, nil
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
		t.opts.TagName = defaultTagName
	}

	return nil
}
