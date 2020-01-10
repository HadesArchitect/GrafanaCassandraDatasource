# C* -> Grafana DS

** EARLY PHASE **

```
create table test (id smallint(6) unsigned NOT NULL, time datetime NOT NULL, temperature smallint(6) NOT NULL);

insert into test (id, time, temperature) values (2, NOW(), 15);

SELECT
  time AS "time",
  temperature
FROM test
WHERE
  $__timeFilter(time)
ORDER BY time
```
