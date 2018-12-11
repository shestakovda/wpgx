package wpgx

import "github.com/pkg/errors"

// ErrConnClosed occurs when an attemtp to use closed conncection
var ErrConnClosed = errors.New("connection is closed")

// ErrUnknownType occurs when collector meets unknown shaper type
var ErrUnknownType = errors.New("unknown shaper type")

// Collector is a generic collection. It can use lists, maps or channels inside
//
// NewItem prepares a translator entity for future loads
//
// Collect puts a new translator item into the inside collection
type Collector interface {
	NewItem() Shaper
	Collect(item Shaper) error
}

// Shaper helps to make database model from business model and vice versa
//
// Extrude makes a database model from business data
//
// Receive fills the data from a database model
type Shaper interface {
	Extrude() Translator
	Receive(Translator) error
}

// Translator helps to isolate model from loading procedure
//
// Translate is intended to associate structure fields with their names
type Translator interface {
	Translate(name string) interface{}
}
