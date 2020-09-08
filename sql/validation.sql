SET @blockCount := (SELECT COUNT(`index`) FROM `block`);
SET @blockIndex := (SELECT `block_index` FROM `counter`);

SELECT 'Neo3 Squirrel SQL Validation', 'Result'
UNION ALL
SELECT 'check block count', IF((SELECT COUNT(`index`) FROM `block`) - (SELECT `block_index` FROM `counter`) = 1, 'PASS', 'FAIL')
UNION ALL
SELECT 'check notification count', IF((SELECT SUM(`notifications`) FROM `applicationlog`) - (SELECT COUNT(`id`) FROM `applicationlog_notification`)=0, 'PASS', 'FAIL')
UNION ALL
SELECT 'check transaction count', IF((SELECT SUM(`txs`) FROM `block`) - (SELECT COUNT(`id`) FROM `transaction`)=0, 'PASS', 'FAIL');
