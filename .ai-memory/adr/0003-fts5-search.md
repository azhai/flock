# ADR 0003: Full-Text Search via SQLite FTS5 (no Meilisearch)

## Status
Accepted

## Context
右上角搜索原本只是摆设，需要一个真正的全文检索。候选：独立搜索引擎 Meilisearch、数据库进程内搜索。
团队偏好零部署、无独立服务、数据与既有表强一致。

## Decision
采用 SQLite **FTS5**（`trigram` tokenizer），建 `topic_search` 虚拟表覆盖「标题 + 全部回复正文」，
用 `bm25()` 相关度排序；通过 AFter INSERT/UPDATE/DELETE 触发器与 `topic`/`post` 增量同步，
启动时（`OpenDB -> EnsureSearchIndex`）幂等建表并在索引为空时全量回填。查询走 `/topics?q=`，
短查询或 2 字以内中文（trigram 需 >=3 字节）回退标题 LIKE。

## Consequences
- 零部署：不引入独立服务/额外运维，数据天然当前。
- 中文可用：trigram 支持子串匹配，2 字以内中文回退 LIKE（ceiling：2 字中文不命中正文）。
- 触发器中用 `DELETE WHERE rowid=..` + 重建该话题 body，故按话题聚合正文。
- 注意 goent `RawQueryRowContext(...).Scan()` 不关 rows 会泄漏读锁连接导致 SQLITE_BUSY，一律用 RawQueryContext+Close。