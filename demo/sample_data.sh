#!/usr/bin/env bash

until printf "" 2>>/dev/null >>/dev/tcp/cassandra/9042; do 
    sleep 5;
    echo "Waiting for cassandra...";
done

echo "Creating keyspace and table..."
cqlsh cassandra -u cassandra -p cassandra -e "CREATE KEYSPACE IF NOT EXISTS test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};"
cqlsh cassandra -u cassandra -p cassandra -e "CREATE TABLE IF NOT EXISTS test.test (sensor_id uuid, registered_at timestamp, temperature int, PRIMARY KEY ((sensor_id), registered_at));"


while true; do
    echo "Writing sample data...";
    cqlsh cassandra -u cassandra -p cassandra -e "insert into test.test (sensor_id, registered_at, temperature) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, toTimestamp(now()), $(shuf -i 18-32 -n 1));";
    cqlsh cassandra -u cassandra -p cassandra -e "insert into test.test (sensor_id, registered_at, temperature) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, toTimestamp(now()), $(shuf -i 12-40 -n 1));";
    sleep 5;
done
