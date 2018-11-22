package wpgx

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
	Prepare(name, text string, cols []string) error
	NewDealer() (Dealer, error)
	Close()
}

// Connect method initialize a new connection pool with uri in a connection string format
// Reserve path is for saving args of failed queries. Useful for debug or data restore
func Connect(uri string, options ...func(*Config) error) (Connector, error) {
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
	return
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

func (c *conn) closed() bool {
	return c == nil || c.pool == nil
}

func (c *conn) Prepare(name, text string, cols []string) (err error) {

	if c.closed() {
		return errors.New("connection is closed")
	}

	s := &stmt{
		text: text,
		cols: cols,
	}

	if s.exec, err = c.pool.Prepare(name, text); err != nil {
		return errors.Wrap(err, "preparing statement")
	}

	c.statements[name] = s

	return
}

func (c *conn) NewDealer() (Dealer, error) {

	if c.closed() {
		return nil, errors.New("connection is closed")
	}

	return nil, nil
}

func (c *conn) Close() {

	if c.closed() {
		return
	}

	for name := range c.statements {
		c.pool.Deallocate(name)
	}

	c.pool.Close()
	c.pool = nil
}
