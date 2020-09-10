SELECT 'Neo3 Squirrel SQL Validation', 'Result'
UNION ALL
SELECT 'check block count', IF((SELECT COUNT(`index`) FROM `block`) - (SELECT `block_index` FROM `counter`) = 1, 'PASS', 'FAIL')
UNION ALL
SELECT 'check notification count', IF((SELECT IFNULL((SELECT SUM(`notifications`) FROM `applicationlog`), 0)) - (SELECT COUNT(`id`) FROM `applicationlog_notification`)=0, 'PASS', 'FAIL')
UNION ALL
SELECT 'check transaction count', IF((SELECT IFNULL((SELECT SUM(`txs`) FROM `block`), 0)) - (SELECT COUNT(`id`) FROM `transaction`)=0, 'PASS', 'FAIL');
