SELECT 'Neo3 Squirrel SQL Validation', 'Result'
UNION ALL
SELECT 'check block count', IF(
    (SELECT COUNT(`index`) FROM `block`) -
    (SELECT `block_index` FROM `counter`) = 1
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check notification count', IF(
    (SELECT IFNULL((SELECT SUM(`notifications`) FROM `applicationlog`), 0)) -
    (SELECT COUNT(`id`) FROM `applicationlog_notification`) = 0
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check transaction count', IF(
    (SELECT IFNULL((SELECT SUM(`txs`) FROM `block`), 0)) -
    (SELECT COUNT(`id`) FROM `transaction`) = 0
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check address transfer count', IF(
    !EXISTS(SELECT `cal`.`addr` `address`, `cal`.`transfers` `get`, `address`.`transfers` `want`
           FROM (SELECT `addr`, SUM(cnt) `transfers`
                 FROM (
                          SELECT `from` `addr`, COUNT(`from`) `cnt`
                          FROM `transfer` WHERE `from` != ''
                          GROUP BY `from`
                          UNION ALL
                          SELECT `to` `addr`, COUNT(`to`) `cnt`
                          FROM `transfer` WHERE `to` != ''
                          GROUP BY `to`
                          UNION ALL
                          SELECT `from` `addr`, -COUNT(id) `cnt`
                          FROM `transfer`
                          WHERE `from` != '' AND `from` = `to`
                          GROUP BY `from`
                      ) `tbl`
                 WHERE `addr` <> ''
                 GROUP BY `tbl`.`addr`) `cal`
                    JOIN `address` ON `cal`.`addr` = `address`.`address`
           WHERE `cal`.`transfers` <> `address`.`transfers`)
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check address first tx time', IF(
    !EXISTS(
        SELECT * FROM (SELECT `addr`, MIN(`block_time`) `block_time` FROM (
            SELECT `from` `addr`, MIN(`block_time`) `block_time`
            FROM `transfer`
            WHERE `from` <> ''
            GROUP BY `from`
            UNION ALL
            SELECT `to` `addr`, MIN(`block_time`) `block_time`
            FROM `transfer`
            WHERE `to` <> ''
            GROUP BY `to`
        ) `tbl` GROUP BY `addr`) `cal`
        JOIN `address` ON `cal`.addr=`address`.`address`
        WHERE `cal`.`block_time` <> `address`.first_tx_time
    )
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check address last tx time', IF(
    !EXISTS(
        SELECT * FROM (SELECT `addr`, MAX(`block_time`) `block_time` FROM (
            SELECT `from` `addr`, MAX(`block_time`) `block_time`
            FROM `transfer`
            WHERE `from` <> ''
            GROUP BY `from`
            UNION ALL
            SELECT `to` `addr`, MAX(`block_time`) `block_time`
            FROM `transfer`
            WHERE `to` <> ''
            GROUP BY `to`
        ) `tbl` GROUP BY `addr`) `cal`
        JOIN `address` ON `cal`.addr=`address`.`address`
        WHERE `cal`.`block_time` <> `address`.`last_tx_time`
    )
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check assets', IF(
    !EXISTS(
       SELECT distinct `contract`
       FROM `transfer`
       WHERE `contract` NOT IN (
           SELECT `contract` FROM `asset`
        ) AND `visible`=TRUE
   )
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check addr asset transfers', IF(
    !EXISTS(
        SELECT `cal`.`address`, `cal`.`transfers` `get`, `tbl`.`transfers` `want`
        FROM (SELECT `address`, SUM(`transfers`) `transfers`
              FROM `addr_asset`
              GROUP BY `address`) `cal`
            JOIN (
                SELECT `addr`, SUM(cnt) `transfers`
                 FROM (
                          SELECT `from` `addr`, COUNT(`from`) `cnt`
                          FROM `transfer`
                          WHERE `from` != '' AND `visible` = true
                          GROUP BY `from`
                          UNION ALL
                          SELECT `to` `addr`, COUNT(`to`) `cnt`
                          FROM `transfer`
                          WHERE `to` != '' AND `visible` = true
                          GROUP BY `to`
                          UNION ALL
                          SELECT `from` `addr`, -COUNT(id) `cnt`
                          FROM `transfer`
                          WHERE `from` != '' AND `from` = `to` AND `visible` = true
                          GROUP BY `from`
                      ) `tbl`
                 WHERE `addr` <> ''
                 GROUP BY `tbl`.`addr`
            ) tbl
            ON `cal`.`address` = `tbl`.`addr`
        WHERE `cal`.`transfers` <> `tbl`.transfers
    )
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check address count', IF(
    (SELECT IFNULL((SELECT COUNT(`id`) FROM `address`), 0)) -
    (SELECT `addr_count` FROM `counter`) = 0
, 'PASS', 'FAIL')

UNION ALL

SELECT 'check NEO & GAS transfers total amount balance', IF(
    !EXISTS(
        SELECT @NEO = '0xde5f57d430d3dece511cf975a8d37848cb9e0525',
               @GAS = '0x668e0c1f9d7b70a99dd9e06eadd4c784d641afbc',
               addr_asset.address,
               addr_asset.contract,
               addr_asset.balance,
               aa.balance
        FROM `addr_asset` JOIN (
            SELECT `addr`, `contract`, SUM(`amount`) balance FROM (
                SELECT `from` addr, `contract`, -SUM(amount) amount
                FROM `transfer`
                WHERE `from` != '' AND `contract` IN (@NEO, @GAS)
                GROUP BY `from`, `contract`
                UNION
                SELECT `to` addr, `contract`, SUM(`amount`) amount
                FROM `transfer`
                WHERE `to` != '' AND `contract` IN (@NEO, @GAS)
                GROUP BY `to`, `contract`
            ) a GROUP BY addr, `contract`
        ) aa
        ON `addr_asset`.`address`=aa.`addr` AND `addr_asset`.`contract`=`aa`.`contract`
        WHERE `addr_asset`.`balance` != `aa`.`balance`
    )
, 'PASS', 'FAIL');