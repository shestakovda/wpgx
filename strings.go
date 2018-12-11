package wpgx

import "database/sql"

// Strings is a simple strings collector
// It is useful in one-time tasks or small scripts, when no models is need
type Strings []string

// NewItem is Strings Shaper constructor
func (s *Strings) NewItem() Shaper { return new(stringShaper) }

// Collect is used to add shaper into Strings
func (s *Strings) Collect(item Shaper) error {
	model, ok := item.(*stringShaper)
	if !ok || model == nil {
		return ErrUnknownType
	}
	if model.NullString.Valid {
		*s = append(*s, model.NullString.String)
	}
	return nil
}

type stringShaper struct{ sql.NullString }

func (s *stringShaper) Extrude() Translator               { return s }
func (s *stringShaper) Receive(model Translator) error    { return nil }
func (s *stringShaper) Translate(name string) interface{} { return s }
