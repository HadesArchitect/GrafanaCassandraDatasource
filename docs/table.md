# Table Mode

In addition to TimeSeries mode datasource supports Table mode to draw tables using Cassandra query results. Use `Merge`, `Sort by`, `Organize fields` and other transformations to shape the table in any desirable way.
There are two ways to plot not a whole timeseries but only last(most rescent) values.
1. Inefficient way

In case if table created with default ascending ordering the most recent value is always stored in the end of partition. To retrieve it `ORDER BY` and `LIMIT` clauses must be used in query:
```
SELECT sensor_id, temperature, registered_at, location
FROM test.test
WHERE sensor_id = 99051fe9-6a9c-46c2-b949-38ef78858dd0
AND registered_at > $__timeFrom and registered_at < $__timeTo
ORDER BY registered_at
LIMIT 1
```
Note that `WHERE IN ()` clause could not be used with `ORDER BY`, so query must be duplicated for any additional `sensor_id`.

2. Efficient way

To query the most recent values efficiently ordering must be specified during the table creation:
```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    registered_at timestamp,
    temperature int,
    location text,
    PRIMARY KEY ((sensor_id), registered_at)
) WITH CLUSTERING ORDER BY (registered_at DESC);
```
After that the most recent value will always be stored in the beginning of partition and could be queried with just `LIMIT` clause:
```
SELECT sensor_id, temperature, registered_at, room_name
FROM test.test
WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd0)
AND registered_at > $__timeFrom AND registered_at < $__timeTo
PER PARTITION LIMIT 1
```
Note that `PER PARTITION LIMIT 1` used instead of `LIMIT 1` to query one row for each partition and not just one row total.
