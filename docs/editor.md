# Query Editor

Query Editor unlocks all possibilities of CQL including Used-Defined Functions, aggregations etc. To enable query editor, press "toggle editor mode" button.

Example (using the sample table from the Query Configurator case):

```
SELECT sensor_id, temperature, registered_at, location FROM test.test WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) AND registered_at > $__timeFrom and registered_at < $__timeTo
```

1. Order of fields in the SELECT expression doesn't matter except `ID` field. This field used to distinguish different time series, so it is important to keep it or any other column with low cardinality on the first position.
* **Identifier** - the first property in the SELECT expression should be the ID, something that uniquely identifies the data (e.g. `sensor_id`)
* **Value** - There should be at least one numeric value among returned fields, if query result will be used to draw graph.
* **Timestamp** - There should be one timestamp value, if query result will be used to draw graph.
* There could be any number of additional fields, however be cautious when using multiple numeric fields as they are interpreted as values by grafana and therefore are drawn on TimeSeries graph.
* Any field returned by query is available to use in `Alias` template, e.g. `{{ location }}`. Datasource interpolates such strings and updates graph legend. 
* Datasource will try to keep all the fields, however it is not always possible since cassandra and grafana use different sets of supported types. Unsupported fields will be removed from response.

2. To filter data by time, use `$__timeFrom` and `$__timeTo` placeholders as in the example. The datasource will replace them with time values from the panel. **Notice** It's important to add the placeholders otherwise query will try to fetch data for the whole period of time. Don't try to specify the timeframe on your own, just put the placeholders. It's grafana's job to specify time limits.

![103153625-1fd85280-4792-11eb-9c00-085297802117](https://user-images.githubusercontent.com/1742301/148654522-8e50617d-0ba9-4c5a-a3f0-7badec92e31f.png)
