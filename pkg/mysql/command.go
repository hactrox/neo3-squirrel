package mysql

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/config"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"strings"
)

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func Exec(query string, args ...interface{}) (sql.Result, error) {
	dbReady()

	q := func() (interface{}, error) {
		result, err := dbClient.db.Exec(query, args...)
		return result, err
	}

	result, err := dbClient.wrappedQuery(q)
	if result == nil {
		return nil, err
	}

	return result.(sql.Result), err
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func Query(query string, args ...interface{}) (*sql.Rows, error) {
	dbReady()

	q := func() (interface{}, error) {
		return dbClient.db.Query(query, args...)
	}

	rows, err := dbClient.wrappedQuery(q)
	if rows == nil {
		return nil, err
	}

	return rows.(*sql.Rows), err
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
func QueryRow(query string, args []interface{}, dest ...interface{}) error {
	dbReady()

	q := func() (interface{}, error) {
		row := dbClient.db.QueryRow(query, args...)
		err := row.Scan(dest...)
		return nil, err
	}

	_, err := dbClient.wrappedQuery(q)
	return err
}

// Trans starts a transaction.
func Trans(txFunc func(*sql.Tx) error) (err error) {
	dbReady()

	tx, err := dbClient.db.Begin()
	if err != nil {
		if autoReconnectDisabled {
			return err
		}

		if dbClient.lostConnection() {
			dbClient.reconnect()
			return Trans(txFunc)
		}

		log.Error(err)
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			if strings.Contains(fmt.Sprintf("%s", p), "SQL execution failed:") {
				panic("SQL execution failed")
			} else {
				panic(p)
			}
		} else if err != nil {
			// Error is recorded inside 'txFunc',
			// no need to log it again.
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Error(err)
			}
		}
	}()

	err = txFunc(tx)
	return err
}

func (db *DB) wrappedQuery(qFunc func() (interface{}, error)) (interface{}, error) {
	for {
		re, err := qFunc()
		if err != nil {
			if autoReconnectDisabled {
				return re, err
			}

			if dbClient.lostConnection() {
				db.reconnect()
				// Retry db query after reconnect.
				continue
			}
			// If it's not connection issues, return error to caller.
			return re, err
		}

		return re, nil
	}
}

// Compose composes raw sql fragments into one query string.
func Compose(query []string) (sql string) {
	sql = strings.Join(query, " ")

	if config.DebugSQLMode() {
		coloredSQL := "\n" + color.BLightGreen(strings.Join(query, "\n"))
		log.Debug(coloredSQL)
	}

	return
}
