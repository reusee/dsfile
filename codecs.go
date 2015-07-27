package dsfile

import (
	"encoding/gob"
	"encoding/json"
	"io"

	"github.com/ugorji/go/codec"
)

type Gob struct{}

func (g *Gob) Encode(w io.Writer, obj interface{}) error {
	return gob.NewEncoder(w).Encode(obj)
}

func (g *Gob) Decode(r io.Reader, obj interface{}) error {
	return gob.NewDecoder(r).Decode(obj)
}

type Json struct{}

func (j *Json) Encode(w io.Writer, obj interface{}) error {
	return json.NewEncoder(w).Encode(obj)
}

func (j *Json) Decode(r io.Reader, obj interface{}) error {
	return json.NewDecoder(r).Decode(obj)
}

var cborHandle = new(codec.CborHandle)

type Cbor struct{}

func (c *Cbor) Encode(w io.Writer, obj interface{}) error {
	return codec.NewEncoder(w, cborHandle).Encode(obj)
}

func (c *Cbor) Decode(r io.Reader, obj interface{}) error {
	return codec.NewDecoder(r, cborHandle).Decode(obj)
}
