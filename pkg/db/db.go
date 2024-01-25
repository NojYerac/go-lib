package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"source.rad.af/libs/go-lib/pkg/log"
	"source.rad.af/libs/go-lib/pkg/tracing"
)

func NewDatabase(config *Configuration, opts ...Option) Database {
	o := &options{
		l: log.Nop(),
		t: tracing.TracerForPackage(),
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

// Open returns a db connection
func (db *database) Open(ctx context.Context) error {
	db.o.l.Trace().Msg("open connection")
	conn, err := sqlx.ConnectContext(ctx, db.driver, db.connStr)
	if err != nil {
		return err
	}
	db.conn = conn
	db.data = &data{eq: conn, o: db.o}
	if db.o.h != nil {
		db.o.h.Register("db_check", db.conn.PingContext)
	}
	return initMetrics(db)
}

// Close releases the connection
func (db *database) Close() error {
	db.o.l.Trace().Msg("close connection")
	return db.conn.Close()
}

// Begin starts a transaction
func (db *database) Begin(parentCtx context.Context) (txx Tx, err error) {
	ctx, deferFunc := db.start(parentCtx, "Begin", "", nil, nil)
	defer deferFunc(err)
	db.o.l.Trace().Msg("begin transaction")
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return
	}
	txx = &transaction{
		data: &data{eq: tx, o: db.o},
		tx:   tx,
		o:    db.o,
	}
	return
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

func (d *data) start(
	parentCtx context.Context,
	op, query string,
	args []interface{},
	dest interface{},
) (ctx context.Context, deferFunc func(error)) {
	logger := zerolog.Ctx(parentCtx)
	ctx, span := d.o.t.Start(parentCtx, op)
	start := time.Now()
	deferFunc = func(err error) {
		latentcy := time.Since(start)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
		durationHistogram.Record(ctx, latentcy.Milliseconds(), semconv.DBOperationKey.String(op))
		logger.Err(err).
			Str("sql", query).
			Interface("arguments", args).
			Interface("result", dest).
			Dur("latentcy", latentcy).
			Msg(op)
	}
	return
}

func (d *data) Select(
	parentCtx context.Context,
	dest interface{},
	query string,
	args ...interface{},
) (err error) {
	ctx, deferFunc := d.start(parentCtx, "Select", query, args, dest)
	defer deferFunc(err)
	err = sqlx.SelectContext(ctx, d.eq, dest, query, args...)
	return err
}

func (d *data) Get(
	parentCtx context.Context,
	dest interface{},
	query string,
	args ...interface{},
) (err error) {
	ctx, deferFunc := d.start(parentCtx, "Get", query, args, dest)
	defer deferFunc(err)
	err = sqlx.GetContext(ctx, d.eq, dest, query, args...)
	return err
}

func (d *data) Exec(
	parentCtx context.Context,
	query string,
	args ...interface{},
) (res sql.Result, err error) {
	ctx, deferFunc := d.start(parentCtx, "Exec", query, args, res)
	defer deferFunc(err)
	res, err = d.eq.ExecContext(ctx, query, args...)
	return
}

func (d *data) Query(
	parentCtx context.Context,
	query string,
	args ...interface{},
) (rows *sqlx.Rows, err error) {
	ctx, deferFunc := d.start(parentCtx, "Query", query, args, rows)
	defer deferFunc(err)
	rows, err = d.eq.QueryxContext(ctx, query, args...)
	return
}

// Tx is the interface for managing a transaction
type Tx interface {
	Commit(context.Context) error
	Rollback(context.Context) error
	DataInterface
}

type transaction struct {
	*data
	tx *sqlx.Tx
	o  *options
}

func (t *transaction) Commit(ctx context.Context) (err error) {
	_, deferFunc := t.start(ctx, "Commit", "", nil, nil)
	defer deferFunc(err)
	err = t.tx.Commit()
	return
}

func (t *transaction) Rollback(ctx context.Context) (err error) {
	_, deferFunc := t.start(ctx, "Rollback", "", nil, nil)
	defer deferFunc(err)
	err = t.tx.Rollback()
	return
}
