package mysql

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// IsDuplicateEntryError checks if duplicated record error occurred or not.
func IsDuplicateEntryError(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		return mysqlErr.Number == 1062
	}

	return false
}

// IsRecordNotFoundError checks if no rows was returned or not.
func IsRecordNotFoundError(err error) bool {
	return err == sql.ErrNoRows
}

// CheckIfRowsNotAffected checks if sql execution affects no rows or not.
func CheckIfRowsNotAffected(result sql.Result, query []string) {
	affectedRows, err := result.RowsAffected()
	if err != nil || affectedRows == 0 {
		if err == nil {
			err = fmt.Errorf("\nSQL execution failed:\n%s", strings.Join(query, "\n"))
		}

		log.Panic(color.BRed(err.Error()))
	}
}
