---
'grafana-cassandra-datasource': minor
---

**CQL is now written in a code editor.** The Query Editor and the variable query editor use Grafana's code editor (Monaco) instead of a plain text box, with syntax highlighting, line numbers and an editor that grows with the query. There is no CQL mode, so CQL is highlighted as SQL. Type `$` for autocomplete: the time range macros `$__timeFrom`, `$__timeTo`, `$__unixEpochFrom` and `$__unixEpochTo`, each with a short description of what it is replaced with, Grafana's `$__from` and `$__to`, and the variables of the dashboard. `Ctrl+Enter` (`Cmd+Enter` on macOS) runs the query, or applies the variable query; leaving the editor still does so too. Saved queries open unchanged and are stored in the same format as before (see [Query Editor](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/editor.md)).

**Incomplete queries no longer show an error.** A new panel used to fail with "Skipping query execution while not all … fields are filled" until every field was set, and only the panel's first query was checked, so one unfinished query could block or break the others. Each query is now checked on its own: queries still being written (empty CQL in the Query Editor; a missing keyspace, table, column or ID value in the Query Configurator) and hidden queries are simply not run, while the complete queries of the same panel run as usual.

**Keyspaces, tables and columns are listed in alphabetical order** in the Query Configurator dropdowns, ignoring case and with numbers in numeric order (`table2` before `table10`), instead of the order Cassandra happened to return them in. System keyspaces are still listed.
