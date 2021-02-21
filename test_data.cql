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
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(65), 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(70), 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(75), 24);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(80), 22);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(85), 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(90), 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(95), 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(100), 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(105), 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(110), 21);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(115), 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, minutesAgo(120), 17);
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
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(60), 13);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(65), 12);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(70), 11);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(75), 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(80), 9);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(85), 8);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(90), 7);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(95), 6);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(100), 6);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(105), 5);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(110), 5);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(115), 6);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, minutesAgo(120), 8);
