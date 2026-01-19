# Variables

Variables allow you to create dynamic filters for your dashboards. They can be used to query values from your Cassandra database and reference them in your queries.

## Basic Variables

Define variables in dashboard Settings > Variables. Each variable executes a CQL query to fetch available values:

![Variables example](assets/variables.png)

Use variables in your queries by referencing them with `$variable_name` syntax.

## Chained Variables

Variables can reference other variables in their queries, creating a chain of dependent values:

![Chained variables example](assets/chained-variables.png)

In this example, the `location` variable depends on the `zone` variable using `$zone` in its query.

**Important:** Variable order matters. Variables can only reference variables defined **above** them in the list. Lower variables can depend on higher variables, but not vice versa.

## Demo Data

The [`demo/sample_data.sh`](../demo/sample_data.sh) script creates example tables including `test.zones_locations` which demonstrates the chained variables feature with zone and location hierarchies.
