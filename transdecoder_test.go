package kvstructure_test

import (
	"fmt"
	"testing"

	. "github.com/andersnormal/kvstructure"
	mm "github.com/andersnormal/kvstructure/mock"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/mock"
)

func TestTransdecodeStruct(t *testing.T) {
	s := &mm.Mock{}
	s.On("Get", "prefix/foo/description").Return(
		&store.KVPair{
			Key:   "prefix/foo/description",
			Value: []byte("bar"),
		},
		nil,
	)
	s.On("Get", "prefix/foo/condition").Return(
		&store.KVPair{
			Key:   "prefix/foo/condition",
			Value: []byte(fmt.Sprint("true")),
		},
		nil,
	)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTransdecoder(
		TransdecoderWithKV(kv),
		TransdecoderWithPrefix("prefix"),
	)

	tt := new(Test)

	assert.NoError(t, err)

	err = td.Transdecode("foo", tt)
	assert.NoError(t, err)
	assert.Equal(t, true, tt.Cond)
	assert.Equal(t, "bar", tt.Desc)
}

func TestTransdecodeString(t *testing.T) {
	s := &mm.Mock{}
	s.On("Get", "prefix/foo").Return(
		&store.KVPair{
			Key:   "prefix/foo",
			Value: []byte("bar"),
		},
		nil,
	)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTransdecoder(
		TransdecoderWithKV(kv),
		TransdecoderWithPrefix("prefix"),
	)

	var tt string

	assert.NoError(t, err)

	err = td.Transdecode("foo", &tt)
	assert.NoError(t, err)
	assert.Equal(t, "bar", tt)
}

func TestTransdecodeInt(t *testing.T) {
	s := &mm.Mock{}
	s.On("Get", "prefix/foo").Return(
		&store.KVPair{
			Key:   "prefix/foo",
			Value: []byte(fmt.Sprint(999)),
		},
		nil,
	)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTransdecoder(
		TransdecoderWithKV(kv),
		TransdecoderWithPrefix("prefix"),
	)

	var tt int

	assert.NoError(t, err)

	err = td.Transdecode("foo", &tt)
	assert.NoError(t, err)
	assert.Equal(t, 999, tt)
}

func TestTransdecodeFloat32(t *testing.T) {
	s := &mm.Mock{}
	s.On("Get", "prefix/foo").Return(
		&store.KVPair{
			Key:   "prefix/foo",
			Value: []byte(fmt.Sprint(float64(9.999))),
		},
		nil,
	)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTransdecoder(
		TransdecoderWithKV(kv),
		TransdecoderWithPrefix("prefix"),
	)

	var tt float32

	assert.NoError(t, err)

	err = td.Transdecode("foo", &tt)
	assert.NoError(t, err)
	assert.Equal(t, float32(9.999), tt)
}
