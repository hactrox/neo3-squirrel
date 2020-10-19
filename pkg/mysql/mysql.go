package mysql

import (
	"database/sql"
	"neo3-squirrel/config"
	"neo3-squirrel/pkg/reconnector"
	"neo3-squirrel/util/log"
	"strings"
	"time"

	// Import mysql driver
	_ "github.com/go-sql-driver/mysql"
)

var (
	connConfig = map[string]string{
		"charset":         "utf8mb4",
		"parseTime":       "True",
		"loc":             "Local",
		"multiStatements": "True",
	}
)

// DB encapsulates MySQL DB variables.
type DB struct {
	db *sql.DB
}

var (
	dbClient *DB

	autoReconnectDisabled = false
)

// Init opens db connection and checks if the connection is valid.
func Init() {
	connStr := config.GetDbConnStr()
	if connCfg := getConnConfig(); connCfg != "" {
		connStr += "?" + connCfg
	}

	log.Infof("Connect to db: %s", config.GetDBInfo())

	db, err := sql.Open("mysql", connStr)
	if err != nil || db.Ping() != nil {
		if err == nil {
			err = db.Ping()
		}

		log.Fatalf("Failed to connect database: %v", err)
	}

	db.SetConnMaxLifetime(10 * time.Second)
	db.SetMaxIdleConns(0)

	dbClient = &DB{db}
}

// DisableAutoReconnect disables auto-reconnect feature.
func DisableAutoReconnect() {
	autoReconnectDisabled = true
}

func getConnConfig() string {
	var configSlice []string

	for k, v := range connConfig {
		configSlice = append(configSlice, k+"="+v)
	}

	return strings.Join(configSlice, "&")
}

func (db *DB) reconnect() {
	reconnector.Reconnect("Mysql", func() bool {
		return db.db.Ping() == nil && serverAlive()
	})
}

func (db *DB) lostConnection(err error) bool {
	errMsg := err.Error()
	if strings.HasSuffix(errMsg, "Server shutdown in progress") ||
		strings.HasSuffix(errMsg, "invalid connection") {
		log.Debug("server shutdown or connection invalid")
		return true
	}

	return db.db.Ping() != nil || !serverAlive()
}

func serverAlive() bool {
	var timestamp int64
	query := "SELECT SQL_NO_CACHE UNIX_TIMESTAMP(CURTIME())"
	err := QueryRow(query, nil, &timestamp)

	return timestamp > 0 && err == nil
}

func dbReady() {
	if dbClient == nil {
		log.Fatal("Cannot execute sql: please init db client first")
	}
}
