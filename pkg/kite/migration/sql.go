package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/sllt/kite/pkg/kite/infra"
	kiteSql "github.com/sllt/kite/pkg/kite/datasource/sql"
)

const (
	createSQLKiteMigrationsTable = `CREATE TABLE IF NOT EXISTS kite_migrations (
    version BIGINT not null ,
    method VARCHAR(4) not null ,
    start_time TIMESTAMP not null ,
    duration BIGINT,
    constraint primary_key primary key (version, method)
);`

	getLastSQLKiteMigration = `SELECT COALESCE(MAX(version), 0) FROM kite_migrations;`

	insertKiteMigrationRowMySQL = `INSERT INTO kite_migrations (version, method, start_time,duration) VALUES (?, ?, ?, ?);`

	insertKiteMigrationRowPostgres = `INSERT INTO kite_migrations (version, method, start_time,duration) VALUES ($1, $2, $3, $4);`
)

// database/sql is the package imported so named it sqlDS.
type sqlDS struct {
	SQL
}

func (s *sqlDS) apply(m migrator) migrator {
	return sqlMigrator{
		SQL:      s.SQL,
		migrator: m,
	}
}

type sqlMigrator struct {
	SQL

	migrator
}

func (d sqlMigrator) checkAndCreateMigrationTable(c *infra.Container) error {
	if _, err := c.SQL.Exec(createSQLKiteMigrationsTable); err != nil {
		return err
	}

	return d.migrator.checkAndCreateMigrationTable(c)
}

func (d sqlMigrator) getLastMigration(c *infra.Container) (int64, error) {
	var lastMigration int64

	err := c.SQL.QueryRowContext(context.Background(), getLastSQLKiteMigration).Scan(&lastMigration)
	if err != nil {
		return -1, fmt.Errorf("sql: %w", err)
	}

	c.Debugf("SQL last migration fetched value is: %v", lastMigration)

	lm2, err := d.migrator.getLastMigration(c)
	if err != nil {
		return -1, err
	}

	return max(lastMigration, lm2), nil
}

func (d sqlMigrator) commitMigration(c *infra.Container, data transactionData) error {
	switch c.SQL.Dialect() {
	case "mysql", "sqlite":
		err := insertMigrationRecord(data.SQLTx, insertKiteMigrationRowMySQL, data.MigrationNumber, data.StartTime)
		if err != nil {
			return err
		}

		c.Debugf("inserted record for migration %v in kite_migrations table", data.MigrationNumber)

	case "postgres":
		err := insertMigrationRecord(data.SQLTx, insertKiteMigrationRowPostgres, data.MigrationNumber, data.StartTime)
		if err != nil {
			return err
		}

		c.Debugf("inserted record for migration %v in kite_migrations table", data.MigrationNumber)
	}

	// Commit transaction
	if err := data.SQLTx.Commit(); err != nil {
		return err
	}

	return d.migrator.commitMigration(c, data)
}

func insertMigrationRecord(tx *kiteSql.Tx, query string, version int64, startTime time.Time) error {
	_, err := tx.Exec(query, version, "UP", startTime, time.Since(startTime).Milliseconds())

	return err
}

func (d sqlMigrator) beginTransaction(c *infra.Container) transactionData {
	sqlTx, err := c.SQL.Begin()
	if err != nil {
		c.Errorf("unable to begin transaction: %v", err)

		return transactionData{}
	}

	cmt := d.migrator.beginTransaction(c)

	cmt.SQLTx = sqlTx

	c.Debug("SQL Transaction begin successful")

	return cmt
}

func (d sqlMigrator) rollback(c *infra.Container, data transactionData) {
	if data.SQLTx == nil {
		return
	}

	if err := data.SQLTx.Rollback(); err != nil {
		c.Error("unable to rollback transaction: %v", err)
	}

	d.migrator.rollback(c, data)

	c.Fatalf("Migration %v failed and rolled back", data.MigrationNumber)
}
