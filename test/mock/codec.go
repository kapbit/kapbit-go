package mock

import (
	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/ymz-ncnk/mok"
)

// CodecMock is a mock implementation of the eventstore.Codec interface.
type CodecMock struct {
	*mok.Mock
}

// NewCodecMock returns a new instance of CodecMock.
func NewCodecMock() CodecMock {
	return CodecMock{mok.New("Codec")}
}

// --- Function type definitions for mocking ---

type (
	CodecEncodeFn func(event evnt.Event) (evs.EventRecord, error)
	CodecDecodeFn func(record evs.EventRecord) (evnt.Event, error)
)

// --- Register methods ---

func (m CodecMock) RegisterEncode(fn CodecEncodeFn) CodecMock {
	m.Register("Encode", fn)
	return m
}

func (m CodecMock) RegisterNEncode(n int, fn CodecEncodeFn) CodecMock {
	m.RegisterN("Encode", n, fn)
	return m
}

func (m CodecMock) RegisterDecode(fn CodecDecodeFn) CodecMock {
	m.Register("Decode", fn)
	return m
}

func (m CodecMock) RegisterNDecode(n int, fn CodecDecodeFn) CodecMock {
	m.RegisterN("Decode", n, fn)
	return m
}

// --- Interface method implementations ---

func (m CodecMock) Encode(event evnt.Event) (record evs.EventRecord, err error) {
	results, err := m.Call("Encode", mok.SafeVal[evnt.Event](event))
	if err != nil {
		panic(err)
	}
	record, _ = results[0].(evs.EventRecord)
	err, _ = results[1].(error)
	return
}

func (m CodecMock) Decode(record evs.EventRecord) (event evnt.Event, err error) {
	results, err := m.Call("Decode", mok.SafeVal[evs.EventRecord](record))
	if err != nil {
		panic(err)
	}
	event, _ = results[0].(evnt.Event)
	err, _ = results[1].(error)
	return
}
