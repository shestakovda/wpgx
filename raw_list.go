package wpgx

import "database/sql"

type RawList []map[string]string

func (s *RawList) NewItem() Shaper { r := make(rawListShaper); return &r }
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
		} else {
			imap[i] = ""
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
