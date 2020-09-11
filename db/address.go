package db

import (
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"strings"
)

var addressInfoColumn = []string{
	"`id`",
	"`address`",
	"`first_tx_time`",
	"`last_tx_time`",
	"`transfers`",
}

// GetAllAddressInfo returns all address info from DB.
func GetAllAddressInfo() []*models.AddressInfo {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(addressInfoColumn, ", ")),
		"FROM `address`",
	}

	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()
	addrInfo := []*models.AddressInfo{}

	for rows.Next() {
		var m models.AddressInfo

		err := rows.Scan(
			&m.ID,
			&m.Address,
			&m.FirstTxTime,
			&m.LastTxTime,
			&m.Transfers,
		)

		if err != nil {
			log.Error(mysql.Compose(query))
			log.Panic(err)
		}

		addrInfo = append(addrInfo, &m)
	}

	return addrInfo
}
