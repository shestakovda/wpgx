package wpgx

import "database/sql"

// Ints is a simple ints (like ids) collector
// It is useful in one-time tasks or small scripts, when no models is need
type Ints []int

// NewItem is Ints Shaper constructor
func (i *Ints) NewItem() Shaper { return new(intShaper) }

// Collect is used to add shaper into Ints
func (i *Ints) Collect(item Shaper) error {
	model, ok := item.(*intShaper)
	if !ok || model == nil {
		return ErrUnknownType
	}
	if model.NullInt64.Valid {
		*i = append(*i, int(model.NullInt64.Int64))
	}
	return nil
}

type intShaper struct{ sql.NullInt64 }

func (i *intShaper) Extrude() Translator               { return i }
func (i *intShaper) Receive(model Translator) error    { return nil }
func (i *intShaper) Translate(name string) interface{} { return i }
