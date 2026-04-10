package store

import evnt "github.com/kapbit/kapbit-go/event"

type Codec interface {
	Encode(event evnt.Event) (record EventRecord, err error)
	Decode(record EventRecord) (event evnt.Event, err error)
}
