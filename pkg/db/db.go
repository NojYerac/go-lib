package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // blank import to allow postgres driver
	"github.com/rs/zerolog"
)

type Configuration struct {
	Driver    string `config:"datbase_driver" validate:"required"`
	DBConnStr string `config:"database_connection_string" validate:"required"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Driver: "postgres",
	}
}

type Option func(*options)

func WithLogger(l zerolog.Logger) Option {
	return func(o *options) {
		o.l = l
	}
}

type options struct {
	l zerolog.Logger
}

func NewDatabase(config *Configuration, opts ...Option) Database {
	o := &options{
		l: zerolog.Nop(),
	}
	for _, applyOpt := range opts {
		applyOpt(o)
	}
	return &database{
		driver:  config.Driver,
		connStr: config.DBConnStr,
		o:       o,
	}
}

// DataInterface is the interface for issuing SQL commands to a database
type DataInterface interface {
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
}

type execerQueryer interface {
	sqlx.ExecerContext
	sqlx.QueryerContext
}

type data struct {
	eq execerQueryer
	o  *options
}

func (d *data) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := sqlx.SelectContext(ctx, d.eq, dest, query, args...)
	d.o.l.Trace().Err(err).Str("sql", query).Interface("args", args).Interface("dest", dest).Dur("latentcy", time.Since(start)).Msg("select")
	return err
}

func (d *data) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := sqlx.GetContext(ctx, d.eq, dest, query, args...)
	d.o.l.Trace().Err(err).Str("sql", query).Interface("args", args).Interface("dest", dest).Dur("latentcy", time.Since(start)).Msg("get")
	return err
}

func (d *data) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := d.eq.ExecContext(ctx, query, args...)
	d.o.l.Trace().Err(err).Str("sql", query).Interface("args", args).Dur("latentcy", time.Since(start)).Msg("exec")
	return res, err
}
func (d *data) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	var _ = zerolog.DurationFieldUnit
	start := time.Now()
	rows, err := d.eq.QueryxContext(ctx, query, args...)
	d.o.l.Trace().Err(err).Str("sql", query).Interface("args", args).Dur("latentcy", time.Since(start)).Msg("query")
	return rows, err
}

// Tx is the interface for managing a transaction
type Tx interface {
	Commit() error
	Rollback() error
	DataInterface
}

type transaction struct {
	*data
	tx *sqlx.Tx
	o  *options
}

func (t *transaction) Commit() error {
	t.o.l.Trace().Msg("commit operation")
	return t.tx.Commit()
}

func (t *transaction) Rollback() error {
	t.o.l.Trace().Msg("rollback operation")
	return t.tx.Rollback()
}

// Database is the interface for managing a database connection
type Database interface {
	Open(context.Context) error
	Close() error
	Begin(context.Context) (Tx, error)
	DataInterface
}

type database struct {
	*data
	conn    *sqlx.DB
	driver  string
	connStr string
	o       *options
}

// Begin starts a transaction
func (db *database) Begin(ctx context.Context) (Tx, error) {
	db.o.l.Trace().Msg("begin transaction")
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &transaction{
		data: &data{eq: tx, o: db.o},
		tx:   tx,
		o:    db.o,
	}, nil
}

// Open returns a db connection
func (db *database) Open(ctx context.Context) error {
	db.o.l.Trace().Msg("open connection")
	conn, err := sqlx.ConnectContext(ctx, db.driver, db.connStr)
	if err != nil {
		return err
	}
	db.conn = conn
	db.data = &data{eq: conn, o: db.o}
	return nil
}

// Close releases the connection
func (db *database) Close() error {
	db.o.l.Trace().Msg("close connection")
	return db.conn.Close()
}
