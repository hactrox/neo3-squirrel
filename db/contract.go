package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strings"
)

var contractStateColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`state`",
	"`contract_id`",
	"`hash`",
	"`name`",
	"`symbol`",
	"`decimals`",
	"`total_supply`",
	"`script`",
	"`manifest`",
}

var contractColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`hash`",
	"`state`",
	"`new_hash`",
	"`contract_id`",
	"`script`",
	"`manifest`",
}

// GetLastContractStateRecord returns the last record of contract state.
func GetLastContractStateRecord() *models.ContractState {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(contractStateColumns, ", ")),
		"FROM `contract_state`",
		"ORDER BY `block_index` DESC, `id` DESC",
		"LIMIT 1",
	}

	var m models.ContractState
	var totalSupplyStr string

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&m.ID,
		&m.BlockIndex,
		&m.BlockTime,
		&m.TxID,
		&m.State,
		&m.ContractID,
		&m.Hash,
		&m.Name,
		&m.Symbol,
		&m.Decimals,
		&totalSupplyStr,
		&m.Script,
		&m.Manifest,
	)

	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	m.TotalSupply = convert.ToDecimal(totalSupplyStr)

	return &m
}

// GetAllContractStatesGroupedByBlockIndex returns all
// contract states from db grouped by block index.
func GetAllContractStatesGroupedByBlockIndex() [][]*models.ContractState {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(contractStateColumns, ", ")),
		"FROM `contract_state`",
		"ORDER BY `block_index` ASC, `id` ASC",
	}

	csList := []*models.ContractState{}
	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	for rows.Next() {
		var m models.ContractState
		var totalSupplyStr string

		err := rows.Scan(
			&m.ID,
			&m.BlockIndex,
			&m.BlockTime,
			&m.TxID,
			&m.State,
			&m.ContractID,
			&m.Hash,
			&m.Name,
			&m.Symbol,
			&m.Decimals,
			&totalSupplyStr,
			&m.Script,
			&m.Manifest,
		)

		if err != nil {
			log.Error(mysql.Compose(query))
			log.Panic(err)
		}

		m.TotalSupply = convert.ToDecimal(totalSupplyStr)
		csList = append(csList, &m)
	}

	len := len(csList)
	if len == 0 {
		return nil
	}

	groupedIndex := 0
	grouped := [][]*models.ContractState{{}}

	// Grouped by block index.
	for i := 0; i < len; i++ {
		cs := csList[i]
		grouped[groupedIndex] = append(grouped[groupedIndex], cs)
		if i+1 >= len || csList[i+1].BlockIndex != cs.BlockIndex {
			groupedIndex++
			grouped = append(grouped, []*models.ContractState{})
		}
	}

	return grouped
}

// HandleContractStates inserts new contract into DB.
func HandleContractStates(contracts, added, deleted []*models.ContractState, migrated map[*models.ContractState]*models.ContractState) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := insertContracStates(sqlTx, contracts); err != nil {
			return err
		}
		if err := updateContracts(sqlTx, added, deleted, migrated); err != nil {
			return err
		}

		return nil
	})
}

func insertContracStates(sqlTx *sql.Tx, contracts []*models.ContractState) error {
	if len(contracts) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `contract_state` (%s)", strings.Join(contractStateColumns[1:], ", ")))

	strBuilder.WriteString("VALUES")

	// Construct (?, ?, ?) list.
	statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(contractStateColumns[1:]))[1:])
	strBuilder.WriteString(strings.Repeat(statement, len(contracts))[1:])

	// Construct sql query args.
	args := []interface{}{}
	for _, c := range contracts {
		args = append(args,
			c.BlockIndex,
			c.BlockTime,
			c.TxID,
			c.State,
			c.ContractID,
			c.Hash,
			c.Name,
			c.Symbol,
			c.Decimals,
			convert.BigFloatToString(c.TotalSupply),
			c.Script,
			c.Manifest,
		)
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query, args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

func updateContracts(sqlTx *sql.Tx, added, deleted []*models.ContractState, migrated map[*models.ContractState]*models.ContractState) error {
	if err := insertContracts(sqlTx, added); err != nil {
		return err
	}

	if len(deleted) > 0 {
		if err := insertContracts(sqlTx, deleted); err != nil {
			return err
		}

		if err := deleteContracts(sqlTx, deleted); err != nil {
			return err
		}

		if err := deleteAddrAssets(sqlTx, deleted); err != nil {
			return err
		}

		deletedContractHashes := []string{}
		for _, cs := range deleted {
			deletedContractHashes = append(deletedContractHashes, cs.Hash)
		}
		if err := deleteAsset(sqlTx, deletedContractHashes); err != nil {
			return err
		}
	}

	if len(migrated) > 0 {
		if err := migrateContracts(sqlTx, migrated); err != nil {
			return err
		}

		if err := hideObseletedTransfers(sqlTx, migrated); err != nil {
			return err
		}

		if err := migrateAddrAssets(sqlTx, migrated); err != nil {
			return err
		}

		if err := disableObseletedAssets(sqlTx, migrated); err != nil {
			return err
		}
	}

	return nil
}

func insertContracts(sqlTx *sql.Tx, added []*models.ContractState) error {
	if len(added) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `contract` (%s)", strings.Join(contractColumns[1:], ", ")))

	strBuilder.WriteString("VALUES")

	// Construct (?, ?, ?) list.
	statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(contractColumns[1:]))[1:])
	strBuilder.WriteString(strings.Repeat(statement, len(added))[1:])

	// Construct sql query args.
	args := []interface{}{}
	for _, c := range added {
		args = append(args,
			c.BlockIndex,
			c.BlockTime,
			c.TxID,
			c.Hash,
			c.State,
			c.NewHash,
			c.ContractID,
			c.Script,
			c.Manifest,
		)
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query, args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

func deleteContracts(sqlTx *sql.Tx, deleted []*models.ContractState) error {
	if len(deleted) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}

	for _, c := range deleted {
		sql := []string{
			"UPDATE `contract`",
			"SET `state` = 'Deleted'",
			fmt.Sprintf("WHERE `hash` = '%s'", c.Hash),
			"LIMIT 1;",
		}

		strBuilder.WriteString(strings.Join(sql, ", "))
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query)
	if err != nil {
		log.Error(err)
	}

	return err
}

func deleteAddrAssets(sqlTx *sql.Tx, deleted []*models.ContractState) error {
	if len(deleted) == 0 {
		return nil
	}

	hashesParam := ""
	for _, cs := range deleted {
		hashesParam += fmt.Sprintf(", '%s'", cs.Hash)
	}
	hashesParam = hashesParam[2:]

	query := []string{
		"DELETE FROM `addr_asset`",
		fmt.Sprintf("WHERE `contract` IN (%s)", hashesParam),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

func migrateContracts(sqlTx *sql.Tx, migrates map[*models.ContractState]*models.ContractState) error {
	if len(migrates) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}

	for new, old := range migrates {
		sql := []string{
			"UPDATE `contract`",
			"SET `state` = 'Migrated'",
			fmt.Sprintf(", `new_hash` = '%s'", new.Hash),
			fmt.Sprintf("WHERE `hash` = '%s'", old.Hash),
			"LIMIT 1",
		}

		strBuilder.WriteString(strings.Join(sql, " "))
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query)
	if err != nil {
		log.Error(err)
	}

	return err
}

func disableObseletedAssets(sqlTx *sql.Tx, migrates map[*models.ContractState]*models.ContractState) error {
	if len(migrates) == 0 {
		return nil
	}

	contracts := ""
	for _, old := range migrates {
		contracts += fmt.Sprintf(", '%s'", old.Hash)
	}
	contracts = contracts[2:]

	query := []string{
		"UPDATE `asset`",
		"SET `destroyed` = TRUE",
		fmt.Sprintf("WHERE `contract` IN (%s)", contracts),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}

func migrateAddrAssets(sqlTx *sql.Tx, migrates map[*models.ContractState]*models.ContractState) error {
	if len(migrates) == 0 {
		return nil
	}

	strBuilder := strings.Builder{}

	for new, old := range migrates {
		update := []string{
			"UPDATE `addr_asset`",
			fmt.Sprintf("SET `contract` = '%s'", new.Hash),
			fmt.Sprintf("WHERE `contract` = '%s'", old.Hash),
		}

		strBuilder.WriteString(fmt.Sprintf("%s;", strings.Join(update, " ")))
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query)
	if err != nil {
		log.Error(err)
	}

	return err
}

func hideObseletedTransfers(sqlTx *sql.Tx, migrates map[*models.ContractState]*models.ContractState) error {
	if len(migrates) == 0 {
		return nil
	}

	var hashes []string
	for _, old := range migrates {
		hashes = append(hashes, fmt.Sprintf("'%s'", old.Hash))
	}

	query := []string{
		"UPDATE `transfer`",
		"SET `visible` = FALSE",
		fmt.Sprintf("WHERE `contract` IN (%s)", strings.Join(hashes, ", ")),
	}

	_, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
	}

	return err
}
