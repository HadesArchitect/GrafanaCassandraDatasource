CREATE KEYSPACE IF NOT EXISTS test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};

CREATE TABLE IF NOT EXISTS test.test (
    id uuid,
    created_at timestamp,
    value int,
    PRIMARY KEY ((id), created_at)
);

insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-21', 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-22', 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-23', 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-24', 24);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-25', 22);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-26', 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-27', 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-28', 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-29', 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-30', 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2021-12-31', 21);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-01', 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-02', 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-03', 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-04', 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-05', 24);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-06', 22);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-07', 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-08', 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-09', 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-10', 19);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-11', 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-12', 21);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-13', 20);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, '2022-01-14', 17);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-15', 15);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-16', 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-17', 18);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-18', 16);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-19', 14);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-20', 12);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-21', 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-22', 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-23', 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-24', 11);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-25', 12);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-26', 14);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-27', 13);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-28', 12);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-29', 11);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-01-30', 10);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-01', 9);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-02', 8);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-03', 7);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-04', 6);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-05', 6);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-06', 5);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-07', 5);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-08', 6);
insert into test.test (id, created_at, value) values (99051fe9-6a9c-46c2-b949-38ef78858dd1, '2022-02-09', 8);
