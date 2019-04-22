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
	Desc string `kvstructure:"description"`
	Cond bool   `kvstructure:"condition"`
}

func TestTranscodeStruct(t *testing.T) {
	s := &mm.Mock{}
	s.On("Put", "foo/description", []byte("bar"), mock.Anything).Return(nil)
	s.On("Put", "foo/condition", []byte(fmt.Sprint("true")), mock.Anything).Return(nil)

	kv, _ := mm.New(s, []string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix("foo"),
	)

	tt := &Test{
		Desc: "bar",
		Cond: true,
	}

	assert.NoError(t, err)

	err = td.Transcode(&tt)
	assert.NoError(t, err)
}
