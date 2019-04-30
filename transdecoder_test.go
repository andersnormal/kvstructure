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
	s.On("Get", "prefix/foo/withomit").Return(
		&store.KVPair{
			Key:   "prefix/foo/withomit",
			Value: []byte{},
		},
		nil,
	)
	s.On("Get", "prefix/foo/proto").Return(
		&store.KVPair{
			Key:   "prefix/foo/proto",
			Value: []byte(""),
		},
		nil,
	)
	s.On("List", "prefix/foo/tests").Return(
		[]*store.KVPair{
			&store.KVPair{
				Key:   "prefix/foo/tests/0",
				Value: []byte(fmt.Sprint("foo")),
			},
		},
		nil,
	)
	s.On("Get", "prefix/foo/tests/0/proto").Return(
		&store.KVPair{
			Key:   "prefix/foo/tests/0/proto",
			Value: []byte("bar"),
		},
		nil,
	)
	s.On("Get", "prefix/foo/tests/0/description").Return(
		&store.KVPair{
			Key:   "prefix/foo/tests/0/description",
			Value: []byte(""),
		},
		nil,
	)
	s.On("Get", "prefix/foo/tests/0/condition").Return(
		&store.KVPair{
			Key:   "prefix/foo/tests/0/condition",
			Value: []byte(fmt.Sprint("true")),
		},
		nil,
	)
	s.On("Get", "prefix/foo/tests/0/withomit").Return(
		&store.KVPair{
			Key:   "prefix/foo/tests/0/withomit",
			Value: []byte{},
		},
		nil,
	)
	s.On("List", "prefix/foo/tests/0/tests").Return(
		[]*store.KVPair{},
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
	assert.Len(t, tt.Tests, 1)
	assert.ElementsMatch(t, tt.Tests, []*Test{&Test{Cond: true, Tests: make([]*Test, 0)}})
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

func TestTransdecodeSlice(t *testing.T) {
	s := &mm.Mock{}
	s.On("List", "prefix/foo").Return(
		[]*store.KVPair{
			&store.KVPair{
				Key:   "prefix/foo/0",
				Value: []byte(fmt.Sprint("foo")),
			},
			&store.KVPair{
				Key:   "prefix/foo/1",
				Value: []byte(fmt.Sprint("bar")),
			},
		},
		nil,
	)

	s.On("Get", "prefix/foo/0").Return(
		&store.KVPair{
			Key:   "prefix/foo/0",
			Value: []byte(fmt.Sprint("foo")),
		},
		nil,
	)

	s.On("Get", "prefix/foo/1").Return(
		&store.KVPair{
			Key:   "prefix/foo/1",
			Value: []byte(fmt.Sprint("bar")),
		},
		nil,
	)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTransdecoder(
		TransdecoderWithKV(kv),
		TransdecoderWithPrefix("prefix"),
	)

	var tt []string

	assert.NoError(t, err)

	err = td.Transdecode("foo", &tt)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"foo", "bar"}, tt)
}
