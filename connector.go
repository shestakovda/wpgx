package wpgx

import "crypto/sha1"
import "encoding/hex"
import "github.com/jackc/pgx"
import "github.com/pkg/errors"

// Connector is the main database connection manager
// As a Dealer it can execute queries in a default transaction
//
// NewDealer spawns new dealer on the street. It needs to be jailed (closed)
//
// Prepare saves query for further execution
//
// Close closes all free dealers with rollback
type Connector interface {
	Dealer
	Prepare(text string, cols []string) (key string, err error)
	NewDealer() (Dealer, error)
	Close()
}

// Connect method initialize a new connection pool with uri in a connection string format
// Reserve path is for saving args of failed queries. Useful for debug or data restore
func Connect(uri string, options ...func(*Config) error) (Connector, error) {
	var err error

	c := new(conn)
	cfg := new(Config)

	if cfg.ConnPoolConfig.ConnConfig, err = pgx.ParseConnectionString(uri); err != nil {
		return nil, errors.Wrap(err, "parsing connection string")
	}

	for i := range options {
		if err = options[i](cfg); err != nil {
			return nil, errors.Wrap(err, "applying connection options")
		}
	}

	if c.pool, err = pgx.NewConnPool(cfg.ConnPoolConfig); err != nil {
		return nil, errors.Wrap(err, "creating connection pool")
	}

	c.statements = make(map[string]*stmt)
	c.reservePath = cfg.ReservePath
	return c, nil
}

type stmt struct {
	text string
	cols []string
	exec *pgx.PreparedStatement
}

type conn struct {
	pool        *pgx.ConnPool
	statements  map[string]*stmt
	reservePath string
}

func (c *conn) ready() error {
	if c == nil || c.pool == nil {
		return ErrConnClosed
	}
	return nil
}

func (c *conn) Prepare(text string, cols []string) (key string, err error) {
	const emsg = "preparing statement"

	if err = c.ready(); err != nil {
		return "", errors.Wrap(err, emsg)
	}

	s := &stmt{
		text: text,
		cols: cols,
	}

	sum := sha1.Sum([]byte(text))
	key = hex.EncodeToString(sum[:])

	if s.exec, err = c.pool.Prepare(key, text); err != nil {
		return "", errors.Wrap(err, emsg)
	}

	c.statements[key] = s
	return
}

func (c *conn) NewDealer() (Dealer, error) {
	var err error
	const emsg = "creating dealer"

	if err := c.ready(); err != nil {
		return nil, errors.Wrap(err, emsg)
	}

	d := &tx{c: c}
	d.Tx, err = c.pool.Begin()
	return d, errors.Wrap(err, emsg)
}

func (c *conn) Deal(result Collector, query string, args ...interface{}) (err error) {
	var d Dealer
	const emsg = "executing query"

	if d, err = c.NewDealer(); err != nil {
		return errors.Wrap(err, emsg)
	}
	defer func() { d.Jail(err == nil) }()

	return d.Deal(result, query, args...)
}

func (c *conn) Load(item Shaper, query string, args ...interface{}) (err error) {
	var d Dealer
	const emsg = "loading item"

	if d, err = c.NewDealer(); err != nil {
		return errors.Wrap(err, emsg)
	}
	defer func() { d.Jail(err == nil) }()

	return d.Load(item, query, args...)
}

func (c *conn) Save(item Shaper, query string, result Collector) (err error) {
	var d Dealer
	const emsg = "saving item"

	if d, err = c.NewDealer(); err != nil {
		return errors.Wrap(err, emsg)
	}
	defer func() { d.Jail(err == nil) }()

	return d.Save(item, query, result)
}

func (c *conn) Jail(commit bool) error { return nil }

func (c *conn) Close() {

	if err := c.ready(); err != nil {
		return
	}

	for name := range c.statements {
		c.pool.Deallocate(name)
	}

	c.pool.Close()
	c.pool = nil
}
