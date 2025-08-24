# Cassandra fat partitions

Cassandra stores data in `partitions` which are minimal storage units for the DB. It means that using the example table

```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    registered_at timestamp,
    temperature int,
    PRIMARY KEY ((sensor_id), registered_at)
);
```

will lead to partitions bloating and performance degradation, because all the data for all time for specific `sensor_id` is stored in just one partition(first part of `PRIMARY KEY` is `PARTITION KEY`).
To avoid that there is a technique called `bucketing`, which basically means that partitions are split up into smaller pieces.
For instance, we can split that example table partitions by time: year, month, day, or even hour and less. What to choose depends on how
much data stored in each partition. To achieve that the example table has to be modified like this:
```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    date date,
    registered_at timestamp,
    temperature int,
    PRIMARY KEY ((sensor_id, date), registered_at)
);
```
After that change the database schema became more effective because of bucketing by date, and queries will have a form of
```
SELECT sensor_id, temperature, registered_at
FROM temperature
WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) 
AND date = '${__from:date:YYYY-MM-DD}'
AND registered_at > $__timeFrom 
AND registered_at < $__timeTo
```
Note that `$__from`/`$__to` variables are used. They are [grafana built-in variables](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__from-and-__to), and they have formatting capabilities which are perfect for our case.
In case when time range includes more than one day, each day has to be added into `AND date IN (...)` predicate. Another way to make it more convenient is to consider using larger buckets, e.g. month instead of day-size.
