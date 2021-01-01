package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"sort"
	"strings"
)

var addressInfoColumn = []string{
	"`id`",
	"`address`",
	"`first_tx_time`",
	"`last_tx_time`",
}

// GetAllAddresses returns all addresses from DB.
func GetAllAddresses() []string {
	query := []string{
		"SELECT `address`",
		"FROM `address`",
	}

	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()
	addresses := []string{}

	for rows.Next() {
		var addr string
		err := rows.Scan(&addr)

		if err != nil {
			log.Error(mysql.Compose(query))
			log.Panic(err)
		}

		addresses = append(addresses, addr)
	}

	return addresses
}

func updateAddressInfo(sqlTx *sql.Tx, delta map[string]*models.AddressInfo) error {
	if len(delta) == 0 {
		return nil
	}

	// Sort addresses to avoid potential sql dead lock.
	addresses := []string{}
	for addr := range delta {
		addresses = append(addresses, addr)
	}

	sort.Strings(addresses)

	var insertionStrBuilder strings.Builder
	var updatesStrBuilder strings.Builder

	addrAdded := uint(0)
	for _, addr := range addresses {
		firstTxTime := delta[addr].FirstTxTime
		lastTxTime := delta[addr].LastTxTime

		// The new address info should be inserted.
		if firstTxTime == lastTxTime {
			addrAdded++
			insertionStrBuilder.WriteString(fmt.Sprintf(", ('%s', %d, %d)",
				addr, firstTxTime, lastTxTime))
			continue
		}

		// The address record should be updated.
		updateSQL := []string{
			"UPDATE `address`",
			fmt.Sprintf("SET `last_tx_time` = %d", lastTxTime),
			fmt.Sprintf("WHERE `address` = '%s'", addr),
			"LIMIT 1",
		}

		updatesStrBuilder.WriteString(strings.Join(updateSQL, " ") + ";")
	}

	sql := ""
	if insertionStrBuilder.Len() > 0 {
		sql += fmt.Sprintf("INSERT INTO `address`(%s) VALUES ", strings.Join(addressInfoColumn[1:], ", "))
		sql += insertionStrBuilder.String()[2:] + ";"
	}
	sql += updatesStrBuilder.String()

	if len(sql) == 0 {
		return nil
	}

	_, err := sqlTx.Exec(sql)
	if err != nil {
		log.Error(sql)
		log.Panic(err)
	}

	if addrAdded > 0 {
		if err := updateAddressCounter(sqlTx, addrAdded); err != nil {
			return err
		}
	}

	return err
}
