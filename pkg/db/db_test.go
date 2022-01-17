package db_test

import (
	"context"
	"database/sql"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "source.rad.af/libs/go-lib/pkg/db"
	"source.rad.af/libs/go-lib/pkg/log"
)

type testRow struct {
	ID int `db:"id"`
}

var _ = Describe("Database", func() {
	var (
		conn       Database
		opCtx, ctx context.Context
		cancel     context.CancelFunc
	)
	var _, sqlMock, err = sqlmock.NewWithDSN("testDB", sqlmock.MonitorPingsOption(true))
	if err != nil {
		panic(err)
	}
	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		config := NewConfiguration()
		config.Driver = "sqlmock"
		config.DBConnStr = "testDB"
		l := log.NewLogger(log.TestConfig)
		opCtx = l.WithContext(ctx)

		sqlMock.ExpectPing()

		conn = NewDatabase(config, WithLogger(l))
		Expect(conn.Open(ctx)).To(Succeed())
	})
	AfterEach(func() {
		Expect(sqlMock.ExpectationsWereMet()).To(Succeed())

		cancel()
	})
	Describe("Select", func() {
		var val []testRow
		BeforeEach(func() {
			sqlMock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
			err = conn.Select(opCtx, &val, "SELECT")
		})
		It("should return a stub connection", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(HaveLen(1))
			Expect(val[0].ID).To(Equal(99))
		})
	})
	Describe("Get", func() {
		var val testRow
		BeforeEach(func() {
			sqlMock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
			err = conn.Get(opCtx, &val, "SELECT")
		})
		It("should return a stub connection", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(val.ID).To(Equal(99))
		})
	})
	Describe("Exec", func() {
		var res sql.Result
		BeforeEach(func() {
			sqlMock.ExpectExec("SELECT").WillReturnResult(sqlmock.NewResult(99, 1))
			res, err = conn.Exec(opCtx, "SELECT")
		})
		It("should return a stub connection", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.LastInsertId()).To(Equal(int64(99)))
			Expect(res.RowsAffected()).To(Equal(int64(1)))
		})
	})
	Describe("Query", func() {
		var (
			rows *sqlx.Rows
			val  testRow
		)
		BeforeEach(func() {
			sqlMock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
			rows, err = conn.Query(opCtx, "SELECT")
			Expect(err).NotTo(HaveOccurred())
			for rows.Next() {
				Expect(rows.StructScan(&val)).To(Succeed())
			}
		})
		It("should return a stub connection", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(val.ID).To(Equal(99))
		})
	})
	Describe("Begin", func() {
		var tx Tx
		BeforeEach(func() {
			sqlMock.ExpectBegin()
			tx, err = conn.Begin(opCtx)
		})
		It("Returns a Tx interface", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(tx).ToNot(BeNil())
		})
		Describe("Commit", func() {
			It("issues a commit", func() {
				sqlMock.ExpectCommit()
				Expect(tx.Commit()).To(Succeed())
			})
		})
		Describe("Rollback", func() {
			It("issues a rollback", func() {
				sqlMock.ExpectRollback()
				Expect(tx.Rollback()).To(Succeed())
			})
		})
	})
	Describe("Close", func() {
		It("closes", func() {
			sqlMock.ExpectClose()
			Expect(conn.Close()).To(Succeed())
		})
	})
})
