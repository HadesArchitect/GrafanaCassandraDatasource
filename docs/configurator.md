# Query Configurator

![Query Configurator](https://user-images.githubusercontent.com/1742301/148654262-b9cb7253-4086-4367-8aae-35ea458fcbb6.png)

Query Configurator is the easiest way to query data. At first, enter the keyspace and table name, then pick proper columns. If keyspace and table names are given correctly, the datasource will suggest the column names automatically.

* **Time Column** - the column storing the timestamp value, it's used to answer "when" question. 
* **Value Column** - the column storing the value you'd like to show. It can be the `value`, `temperature` or whatever property you need.
* **ID Column** - the column to uniquely identify the source of the data, e.g. `sensor_id`, `shop_id` or whatever allows you to identify the origin of data.

After that, you have to specify the `ID Value`, the particular ID of the data origin you want to show. You may need to enable "ALLOW FILTERING" although we recommend to avoid it.

**Example** Imagine you want to visualise reports of a temperature sensor installed in your smart home. Given the sensor reports its ID, time, location and temperature every minute, we create a table to store the data and put some values there:

```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    registered_at timestamp,
    temperature int,
    location text,
    PRIMARY KEY ((sensor_id), registered_at)
);

insert into temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:21:59.001+0000, 18, "kitchen");
insert into temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:22:59.001+0000, 19, "kitchen");
insert into temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:23:59.001+0000, 20, "kitchen");
```

In this case, we have to fill the configurator fields the following way to get the results:

* **Keyspace** - smarthome *(keyspace name)*
* **Table** - temperature *(table name)*
* **Time Column** - registered_at *(occurence)*
* **Value Column** - temperature *(value to show)*
* **ID Column** - sensor_id *(ID of the data origin)*
* **ID Value** - 99051fe9-6a9c-46c2-b949-38ef78858dd0 *ID of the sensor*
* **ALLOW FILTERING** - FALSE *(not required, so we are happy to avoid)*

In case of a few origins (multiple sensors) you will need to add more rows. If your case is as simple as that, query configurator will be a good choice, otherwise  please proceed to the [Query Editor](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/editor.md).