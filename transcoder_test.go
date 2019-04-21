package kvstructure_test

import (
	"testing"

	. "github.com/andersnormal/kvstructure"

	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/mock"
	"github.com/stretchr/testify/assert"
)

type Test struct {
	Desc string `kvstructure:"description"`
}

func TestTranscodeStruct(t *testing.T) {
	s := &Test{
		Desc: "foo",
	}

	kv, _ := mock.New([]string{"localhost"}, &store.Config{})

	td, err := NewTranscoder(
		TranscoderWithKV(kv),
		TranscoderWithPrefix(""),
	)

	assert.NoError(t, err)

	err = td.Transcode(&s)
	assert.NoError(t, err)

	// tests := []struct {
	// 	desc      string
	// 	s         interface{}
	// 	prefix    string
	// 	endpoints []string
	// 	options   *store.Config
	// }{
	// 	{
	// 		desc:      "with boolean",
	// 		endpoints: []string{"localhost"},
	// 		options:   &store.Config{},
	// 		s: struct {
	// 			Desc string `kvstructure:"description"`
	// 		}{
	// 			Desc: "foo",
	// 		},
	// 	},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.desc, func(t *testing.T) {
	// 		kv, _ := mock.New(tt.endpoints, tt.options)

	// 		td, err := NewTranscoder(
	// 			TranscoderWithKV(kv),
	// 			TranscoderWithPrefix(tt.prefix),
	// 		)

	// 		assert.NoError(t, err)

	// 		err = td.Transcode(tt.options)
	// 		assert.NoError(t, err)
	// 	})
	// }
}
