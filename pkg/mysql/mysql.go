package mysql

import (
	"database/sql"
	"neo3-squirrel/config"
	"neo3-squirrel/log"
	"neo3-squirrel/pkg/reconnector"
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

	log.Infof("Connect to db: [%s]", connStr)

	db, err := sql.Open("mysql", connStr)
	if err != nil || db.Ping() != nil {
		if err == nil {
			err = db.Ping()
		}

		log.Fatalf("Failed to connect database: %v", err)
	}

	db.SetConnMaxLifetime(1 * time.Hour)

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
	reconnector.Reconnect("Mysql", func() error {
		return db.db.Ping()
	})
}

func (db *DB) lostConnection() bool {
	return db.db.Ping() != nil
}

func dbReady() {
	if dbClient == nil {
		log.Fatal("Cannot execute sql: please init db client first")
	}
}
