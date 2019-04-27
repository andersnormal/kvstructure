package kvstructure_test

import (
	"fmt"
	"testing"

	. "github.com/andersnormal/kvstructure"
	mm "github.com/andersnormal/kvstructure/mock"

	"github.com/docker/libkv/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Test struct {
	Desc     string `kvstructure:"description,omitempty"`
	Cond     bool   `kvstructure:"condition" json:"peer,omitempty"`
	Proto    string `protobuf:"bytes,1,opt,name=proto,proto3" json:"proto"`
	WithOmit string `json:"with_omit,omitempty"`
	Ignore   string `json:"-"`
}

func TestTranscodeStruct(t *testing.T) {
	s := &mm.Mock{}
	s.On("Put", "prefix/foo/description", []byte("bar"), mock.Anything).Return(nil)
	s.On("Put", "prefix/foo/condition", []byte(fmt.Sprint("true")), mock.Anything).Return(nil)
	s.On("Put", "prefix/foo/proto", []byte(fmt.Sprint("\"\"")), mock.Anything).Return(nil)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix("prefix"),
	)

	tt := &Test{
		Desc: "bar",
		Cond: true,
	}

	assert.NoError(t, err)

	err = td.Transcode("foo", &tt)
	assert.NoError(t, err)
}

func TestTranscodeString(t *testing.T) {
	s := &mm.Mock{}
	s.On("Put", "prefix/foo", []byte("bar"), mock.Anything).Return(nil)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix("prefix"),
	)

	tt := "bar"

	assert.NoError(t, err)

	err = td.Transcode("foo", &tt)
	assert.NoError(t, err)
}

func TestTranscodeInt(t *testing.T) {
	s := &mm.Mock{}
	s.On("Put", "prefix/foo", []byte(fmt.Sprint(999)), mock.Anything).Return(nil)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix("prefix"),
	)

	tt := 999

	assert.NoError(t, err)

	err = td.Transcode("foo", &tt)
	assert.NoError(t, err)
}

func TestTranscodeUint(t *testing.T) {
	s := &mm.Mock{}
	s.On("Put", "prefix/foo", []byte(fmt.Sprint(uint(999))), mock.Anything).Return(nil)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix("prefix"),
	)

	tt := uint(999)

	assert.NoError(t, err)

	err = td.Transcode("foo", &tt)
	assert.NoError(t, err)
}

func TestTranscodeUint64(t *testing.T) {
	s := &mm.Mock{}
	s.On("Put", "prefix/foo", []byte(fmt.Sprint(uint64(999))), mock.Anything).Return(nil)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix("prefix"),
	)

	tt := uint64(999)

	assert.NoError(t, err)

	err = td.Transcode("foo", &tt)
	assert.NoError(t, err)
}

func TestTranscodeFloat32(t *testing.T) {
	s := &mm.Mock{}
	s.On("Put", "prefix/foo", []byte(fmt.Sprint(float64(9.999))), mock.Anything).Return(nil)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix("prefix"),
	)

	tt := float64(9.999)

	assert.NoError(t, err)

	err = td.Transcode("foo", &tt)
	assert.NoError(t, err)
}
