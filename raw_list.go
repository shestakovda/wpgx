package wpgx

import "database/sql"

// RawList is a simple data collector
// It is useful in one-time tasks or small scripts, when no models is need
type RawList []map[string]string

// NewItem is RawList Shaper constructor
func (s *RawList) NewItem() Shaper { r := make(rawListShaper); return &r }

// Collect is used to add shaper into RawList
func (s *RawList) Collect(item Shaper) error {
	model, ok := item.(*rawListShaper)
	if !ok || model == nil {
		return ErrUnknownType
	}
	rmap := *model
	imap := make(map[string]string, len(rmap))
	for i := range rmap {
		if rmap[i].Valid {
			imap[i] = rmap[i].String
		}
	}

	*s = append(*s, imap)
	return nil
}

type rawListShaper map[string]*sql.NullString

func (r *rawListShaper) Extrude() Translator            { return r }
func (r *rawListShaper) Receive(model Translator) error { return nil }
func (r *rawListShaper) Translate(name string) interface{} {
	if _, ok := (*r)[name]; !ok {
		(*r)[name] = new(sql.NullString)
	}
	return (*r)[name]
}
