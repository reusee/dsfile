package dsfile

import (
	crand "crypto/rand"
	"encoding/binary"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

func init() {
	var seed int64
	binary.Read(crand.Reader, binary.LittleEndian, &seed)
	rand.Seed(seed)
}

type File struct {
	obj    interface{}
	path   string
	codec  Codec
	locker sync.Locker
	cbs    chan func()
}

type Codec interface {
	Decode(io.Reader, interface{}) error
	Encode(io.Writer, interface{}) error
}

func New(obj interface{}, path string, codec Codec, locker sync.Locker) (*File, error) {
	// check object
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return nil, makeErr(nil, "object must be a pointer")
	}

	// init
	file := &File{
		obj:    obj,
		path:   path,
		codec:  codec,
		locker: locker,
		cbs:    make(chan func()),
	}

	// try lock
	done := make(chan struct{})
	go func() {
		locker.Lock()
		close(done)
	}()
	select {
	case <-time.NewTimer(time.Second * 1).C:
		return nil, makeErr(nil, "lock fail")
	case <-done:
	}

	// try load from file
	dbFile, err := os.Open(path)
	if err == nil {
		defer dbFile.Close()
		err = codec.Decode(dbFile, obj)
		if err != nil {
			return nil, makeErr(err, "decode error")
		}
	}

	// loop
	go func() {
		for {
			cb, ok := <-file.cbs
			if !ok {
				return
			}
			cb()
		}
	}()

	return file, nil
}

func (f *File) Save() error {
	done := make(chan struct{})
	var err error
	f.cbs <- func() {
		defer close(done)
		tmpPath := f.path + "." + strconv.FormatInt(rand.Int63(), 10) + ".tmp"
		var tmpF *os.File
		tmpF, err = os.Create(tmpPath)
		if err != nil {
			err = makeErr(err, "open temp file")
			return
		}
		defer tmpF.Close()
		err = f.codec.Encode(tmpF, f.obj)
		if err != nil {
			err = makeErr(err, "encode error")
			return
		}
		err = os.Rename(tmpPath, f.path)
		if err != nil {
			err = makeErr(err, "rename temp file")
			return
		}
	}
	<-done
	return err
}

func (f *File) Close() {
	close(f.cbs)
	f.locker.Unlock()
}
