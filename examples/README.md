# Examples

Runnable examples for sqbl, organized from fundamentals to real-world scenarios.

## Cheat Sheet

Start here if you just want a quick reference covering the most common SQL patterns.

| Example | Covers | Summary |
|---------|--------|---------|
| [basic_fundamentals](./basic_fundamentals/) | `SELECT *`<br>`WHERE`<br>`ORDER BY` / `LIMIT`<br>`DISTINCT`<br>`COUNT` / `SUM` / `AVG`<br>`GROUP BY`<br>`INNER JOIN` / `LEFT JOIN`<br>subquery in `WHERE`<br>`$1` / `:name` / `?` | SQL fundamentals in one file |

## basic_

Single-feature examples. Read these when you want to understand one specific operation in depth.

| Example | Covers | Summary |
|---------|--------|---------|
| [basic_select](./basic_select/) | `SELECT`<br>`WHERE`<br>`ORDER BY`<br>`LIMIT` | Simple SELECT, filtering, OR conditions |
| [basic_insert](./basic_insert/) | `INSERT`<br>`VALUES`<br>`ON CONFLICT`<br>`RETURNING` | Single and multi-row insert, upsert |
| [basic_update](./basic_update/) | `UPDATE`<br>`SET`<br>`WHERE`<br>`RETURNING` | Single and multi-column update |
| [basic_delete](./basic_delete/) | `DELETE`<br>`WHERE`<br>`RETURNING` | Delete with conditions |
| [basic_params](./basic_params/) | `$1`<br>`:name`<br>`?` | Indexed, named, and positional bind parameters |
| [basic_ddl](./basic_ddl/) | `CREATE TABLE` / `INDEX`<br>`ALTER TABLE`<br>`DROP TABLE` / `INDEX` | Table definitions, indexes, schema changes |
| [basic_subquery](./basic_subquery/) | subquery in `FROM`<br>subquery in `WHERE`<br>`IN`<br>`UNION` | Subqueries in FROM and WHERE, UNION |
| [basic_cte](./basic_cte/) | `WITH`<br>`WITH RECURSIVE` | Simple CTE, chained CTEs, recursive CTE |
| [basic_window](./basic_window/) | `ROW_NUMBER`<br>`RANK` / `DENSE_RANK`<br>`SUM OVER`<br>`PARTITION BY` | Window functions, running totals |

## pro_

Multi-step scenarios based on real-world use cases.

| Example | Covers | Summary |
|---------|--------|---------|
| [pro_job_queue](./pro_job_queue/) | `FOR UPDATE SKIP LOCKED`<br>`UPDATE`<br>`RETURNING` | Atomic job claiming for concurrent workers |
| [pro_audit_log](./pro_audit_log/) | `INSERT ON CONFLICT DO UPDATE`<br>`RETURNING` | Upsert with audit trail |
| [pro_ranking](./pro_ranking/) | `WITH` (CTE)<br>`RANK() OVER`<br>`INNER JOIN` | Monthly sales ranking per region |
