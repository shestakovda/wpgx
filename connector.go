package wpgx

import "os"
import "path/filepath"
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
func Connect(uri, reserve string, poolsize int, debug bool) (Connector, error) {
	var err error

	c := new(conn)

	if err = c.setReservePath(reserve); err != nil {
		return nil, errors.Wrap(err, "preparing connection")
	}

	pcfg := pgx.ConnPoolConfig{MaxConnections: poolsize}

	if pcfg.ConnConfig, err = pgx.ParseConnectionString(uri); err != nil {
		return nil, errors.Wrap(err, "parsing connection string")
	}

	pcfg.ConnConfig.Logger = new(logger)

	if debug {
		pcfg.ConnConfig.LogLevel = pgx.LogLevelDebug
	} else {
		pcfg.ConnConfig.LogLevel = pgx.LogLevelWarn
	}

	if c.pool, err = pgx.NewConnPool(pcfg); err != nil {
		return nil, errors.Wrap(err, "creating connection pool")
	}

	c.statements = make(map[string]*stmt)

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

func (c *conn) setReservePath(possible string) (err error) {

	if possible == "" {
		return
	}

	var path string

	if path, err = filepath.Abs(possible); err != nil {
		return errors.Wrap(err, "checking reserve path")
	}

	var info os.FileInfo

	if info, err = os.Stat(path); err != nil {
		return errors.Wrap(err, "testing reserve path")
	}

	if !info.IsDir() {
		return errors.New("reserve path is not a directory")
	}

	c.reservePath = path
	return
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
