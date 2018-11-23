package wpgx

import "database/sql"

type Strings []string

func (s *Strings) NewItem() Shaper { return new(stringShaper) }
func (s *Strings) Collect(item Shaper) error {
	model, ok := item.(*stringShaper)
	if !ok {
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
