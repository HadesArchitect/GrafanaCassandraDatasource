# Unix epoch time format

Usually there are no problems - Cassandra can store timestamps using different formats as shown in [documentation](https://cassandra.apache.org/doc/latest/cassandra/cql/types.html#timestamps).
However, it is not always enough. One of possible cases could be unix time, which is just number of seconds or milliseconds and usually stored as integer type.

1. If time is stored as a number of milliseconds in a `bigint` column, then it should be converted into the `timestamp` type before return the data to grafana:
```
SELECT sensor_id, temperature, dateOf(maxTimeuuid(registered_at)), location
FROM test.test WHERE sensor_id = 99051fe9-6a9c-46c2-b949-38ef78858dd0
AND registered_at > $__timeFrom AND registered_at < $__timeTo
```
This query returns proper timestamp even if it stored as number of milliseconds.

2. If time is stored as a number of seconds, then it is not possible to convert it into the timestamp natively, but there is a trick:
```
SELECT sensor_id, temperature, dateOf(maxTimeuuid(registered_at*1000)), location
FROM test.test WHERE sensor_id = 99051fe9-6a9c-46c2-b949-38ef78858dd0
AND registered_at > $__unixEpochFrom AND registered_at < $__unixEpochTo
```
* There are two important parts in this query:
  * `dateOf(maxTimeuuid(registered_at*1000))` used to convert seconds to milliseconds(`registered_at*1000`) and then to convert milliseconds to `timestamp` type, which is handed over to grafana.
  * `$__unixEpochFrom` and `$__unixEpochTo` are variables with unix time in the seconds format that are used to fill out conditions part of the query.