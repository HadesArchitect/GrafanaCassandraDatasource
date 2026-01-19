#!/usr/bin/env bash

until printf "" 2>>/dev/null >>/dev/tcp/cassandra/9042; do 
    sleep 5;
    echo "Waiting for cassandra...";
done

echo "Creating keyspace and tables..."
cqlsh cassandra -e "CREATE KEYSPACE IF NOT EXISTS test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};"
cqlsh cassandra -e "CREATE TABLE IF NOT EXISTS test.test (sensor_id uuid, registered_at timestamp, temperature int, location text, PRIMARY KEY ((sensor_id), registered_at)) WITH CLUSTERING ORDER BY (registered_at DESC);"
cqlsh cassandra -e "CREATE TABLE IF NOT EXISTS test.sensors_locations (bucket text, location text, sensor_id uuid, PRIMARY KEY((bucket), location));"
cqlsh cassandra -e "CREATE TABLE IF NOT EXISTS test.zones_locations (bucket text, zone text, location text, PRIMARY KEY((bucket), zone, location));"

echo "Inserting zone hierarchy data..."
cqlsh cassandra -e "insert into test.zones_locations (bucket, zone, location) values ('default', 'ground_floor', 'kitchen');";
cqlsh cassandra -e "insert into test.zones_locations (bucket, zone, location) values ('default', 'first_floor', 'bedroom');";
cqlsh cassandra -e "insert into test.zones_locations (bucket, zone, location) values ('default', 'ground_floor', 'living_room');";
cqlsh cassandra -e "insert into test.zones_locations (bucket, zone, location) values ('default', 'first_floor', 'bathroom');";

cqlsh cassandra -e "insert into test.sensors_locations (bucket, location, sensor_id) values ('default', 'kitchen', 99051fe9-6a9c-46c2-b949-38ef78858dd0);";
cqlsh cassandra -e "insert into test.sensors_locations (bucket, location, sensor_id) values ('default', 'bedroom', 99051fe9-6a9c-46c2-b949-38ef78858dd1);";
cqlsh cassandra -e "insert into test.sensors_locations (bucket, location, sensor_id) values ('default', 'living_room', 99051fe9-6a9c-46c2-b949-38ef78858dd2);";
cqlsh cassandra -e "insert into test.sensors_locations (bucket, location, sensor_id) values ('default', 'bathroom', 99051fe9-6a9c-46c2-b949-38ef78858dd3);";

echo "Writing sample data for last 5 minutes...";
# Generate data points for last 5 minutes (60 data points at 5-second intervals)
for i in {1..60}; do
    # Calculate timestamp for i*5 seconds ago
    # Using GNU date syntax (appropriate for Docker container)
    timestamp=$(date -u -d "$(($i * 5)) seconds ago" +"%Y-%m-%d %H:%M:%S")
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '${timestamp}', $(shuf -i 18-32 -n 1), 'kitchen');";
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '${timestamp}', $(shuf -i 12-40 -n 1), 'bedroom');";
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd2, '${timestamp}', $(shuf -i 20-26 -n 1), 'living_room');";
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd3, '${timestamp}', $(shuf -i 18-28 -n 1), 'bathroom');";
done

echo "Sample data preload completed.";

while true; do
    echo "Writing sample data...";
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, toTimestamp(now()), $(shuf -i 18-32 -n 1), 'kitchen');";
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, toTimestamp(now()), $(shuf -i 12-40 -n 1), 'bedroom');";
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd2, toTimestamp(now()), $(shuf -i 20-26 -n 1), 'living_room');";
    cqlsh cassandra -e "insert into test.test (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd3, toTimestamp(now()), $(shuf -i 18-28 -n 1), 'bathroom');";
    sleep 5;
done

# cqlsh -e "SELECT cast(location as text), cast(sensor_id as text) from test.sensors_locations where bucket = 'default'"
