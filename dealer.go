package wpgx

import "github.com/jackc/pgx"
import "github.com/pkg/errors"

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
	Deal(result Collector, query string, args ...interface{}) error
	Load(item Shaper, query string, args ...interface{}) error
	Save(item Shaper, query string, result Collector) error
	Jail(commit bool) error
}

type tx struct {
	*pgx.Tx
	c *conn
}

func (t *tx) closed() bool {
	return t == nil || t.Tx == nil || t.c.closed()
}

func (t *tx) Deal(result Collector, query string, args ...interface{}) (err error) {

	if t.closed() {
		return errors.New("connection is closed")
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

		if err = result.Append(item); err != nil {
			return errors.Wrap(err, "collecting item")
		}
	}

	return errors.Wrap(rows.Err(), "checking result")
}

func (t *tx) Load(item Shaper, query string, args ...interface{}) (err error) {

	if t.closed() {
		return errors.New("connection is closed")
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

func (t *tx) Jail(commit bool) (err error) {
	const emsg = "closing transaction"
	if !commit {
		return errors.Wrap(t.Rollback(), emsg)
	}
	defer t.Rollback()
	return errors.Wrap(t.Commit(), emsg)
}
