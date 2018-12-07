package wpgx

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

// Dealer is an active subject, like an opened transaction, for query performing
//
// Deal! It executes query and loads result into a data collector. Pass nil when no result needed
//
// Load gets just one item from the database. When no collection needed
//
// Save inserts item into database. Result may need for getting new ID or properties
//
// Jail (aka Close) ends a transaction with commit or rollback respective to the flag
type Dealer interface {
	Cook(text string, cols ...string) (string, error)
	Deal(result Collector, query string, args ...interface{}) error
	Load(item Shaper, query string, args ...interface{}) error
	Save(item Shaper, key string, result Collector) error
	Jail(commit bool) error
}

type tx struct {
	*pgx.Tx
	c *conn
}

func (t *tx) ready() error {
	if t == nil || t.Tx == nil || t.c == nil {
		return ErrConnClosed
	}
	return t.c.ready()
}

func (t *tx) Cook(text string, cols ...string) (key string, err error) {
	const emsg = "preparing statement"

	if err = t.ready(); err != nil {
		return "", errors.Wrap(err, emsg)
	}

	sum := sha1.Sum([]byte(text))
	key = hex.EncodeToString(sum[:])

	if _, err = t.Tx.Prepare(key, text); err != nil {
		return "", errors.Wrap(err, emsg)
	}

	t.c.Lock()
	t.c.statements[key] = cols
	t.c.Unlock()

	if t.c.reservePath == "" {
		return
	}

	path := filepath.Join(t.c.reservePath, key+".pgsql")
	err = ioutil.WriteFile(path, []byte(text), 0755)
	return key, errors.Wrap(err, emsg)
}

func (t *tx) Deal(result Collector, query string, args ...interface{}) (err error) {

	if err = t.ready(); err != nil {
		return errors.Wrap(err, "executing query")
	}

	if result == nil {
		_, err = t.Exec(query, args...)
		return errors.Wrap(err, "executing query")
	}

	var rows *pgx.Rows

	if rows, err = t.Query(query, args...); err != nil {
		return errors.Wrap(err, "selecting data")
	}
	defer rows.Close()

	names := rows.FieldDescriptions()
	places := make([]interface{}, len(names))

	for rows.Next() {
		item := result.NewItem()

		if item == nil {
			break
		}

		model := item.Extrude()

		for i := range names {
			places[i] = model.Translate(names[i].Name)
		}

		if err = rows.Scan(places...); err != nil {
			return errors.Wrap(err, "scanning data row")
		}

		if err = item.Receive(model); err != nil {
			return errors.Wrap(err, "receiving model")
		}

		if err = result.Collect(item); err != nil {
			return errors.Wrap(err, "collecting item")
		}
	}

	return errors.Wrap(rows.Err(), "checking result")
}

func (t *tx) Load(item Shaper, query string, args ...interface{}) (err error) {

	if err = t.ready(); err != nil {
		return errors.Wrap(err, "loading item")
	}

	var rows *pgx.Rows

	if rows, err = t.Query(query, args...); err != nil {
		return errors.Wrap(err, "selecting data")
	}
	defer rows.Close()

	names := rows.FieldDescriptions()
	places := make([]interface{}, len(names))

	if rows.Next() {
		model := item.Extrude()

		for i := range names {
			places[i] = model.Translate(names[i].Name)
		}

		if err = rows.Scan(places...); err != nil {
			return errors.Wrap(err, "scanning data row")
		}

		if err = item.Receive(model); err != nil {
			return errors.Wrap(err, "receiving model")
		}
	}

	return errors.Wrap(rows.Err(), "checking result")
}

func (t *tx) Save(item Shaper, key string, result Collector) (err error) {

	if err = t.ready(); err != nil {
		return errors.Wrap(err, "saving item")
	}

	t.c.RLock()
	cols, ok := t.c.statements[key]
	t.c.RUnlock()
	if !ok {
		return errors.New("unknown prepared query key: " + key)
	}

	args := make([]interface{}, len(cols))
	model := item.Extrude()

	for i := range cols {
		args[i] = model.Translate(cols[i])
	}

	defer func() {
		if err == nil || t.c.reservePath == "" {
			return
		}
		const emsg = "reserving data"
		const etpl = "\n%+v\nreserve data: %+v"

		dump := make(map[string]interface{})
		for i := range cols {
			dump[cols[i]] = args[i]
		}

		text, ex := json.MarshalIndent(dump, "", "  ")
		if ex != nil {
			fmt.Printf(etpl, errors.Wrap(ex, emsg), dump)
		}

		sum := sha1.Sum([]byte(text))
		hash := hex.EncodeToString(sum[:])
		path := filepath.Join(t.c.reservePath, key+"_"+hash+".json")

		if ex = ioutil.WriteFile(path, text, 0755); ex != nil {
			fmt.Printf(etpl, errors.Wrap(ex, emsg), dump)
		}
	}()

	return t.Deal(result, key, args...)
}

func (t *tx) Jail(commit bool) (err error) {
	const emsg = "closing transaction"
	defer func() {
		t.Rollback()
		t.Tx = nil
		t.c = nil
		t = nil
	}()
	if !commit {
		return errors.Wrap(t.Rollback(), emsg)
	}
	return errors.Wrap(t.Commit(), emsg)
}
