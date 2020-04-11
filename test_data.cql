CREATE KEYSPACE IF NOT EXISTS test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};

CREATE TABLE IF NOT EXISTS test.test (
    id uuid,
    created_at timestamp,
    value int,
    PRIMARY KEY ((id), created_at)
);

CREATE FUNCTION IF NOT EXISTS test.minutesAgo(minutes int) 
  CALLED ON NULL INPUT 
  RETURNS timestamp
  LANGUAGE java AS '
    long now = System.currentTimeMillis();
    if (minutes == null)
      return new Date(now);
    return new Date(now - (minutes.intValue() * 60 * 1000));
  ';

insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(0), 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(5), 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(10), 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(15), 24);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(20), 22);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(25), 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(30), 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(35), 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(40), 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(45), 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(50), 21);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(55), 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(60), 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(0), 15);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(5), 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(10), 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(15), 16);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(20), 14);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(25), 12);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(30), 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(35), 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(40), 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(45), 11);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(50), 12);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(55), 14);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(60), 20);
