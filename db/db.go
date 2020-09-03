package db

import "neo3-squirrel/pkg/mysql"

// Init initializes db layer.
func Init() {
	mysql.Init()
}
