\c template1
drop database if exists contacts;

create database contacts;
GRANT CONNECT ON DATABASE contacts TO advanced;
GRANT ALL PRIVILEGES ON DATABASE contacts TO advanced;
ALTER USER advanced WITH PASSWORD 'set!Application:Password'; -- This is in the config, should be modified before imported then deleted

\c contacts;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;

create table contacts(
 id uuid primary key default uuid_generate_v4(),
 first_name text,
 last_name text,
 phone bigint,
 office_phone bigint,
 city text,
 state text,
 zip text,
 enabled bool default true

);

insert into contacts (first_name, last_name, phone, office_phone, city, state, zip)
values ('Oscar','Torrealba','4126780017','111232','Quibor','Lara','3061');
insert into contacts (first_name, last_name, phone, office_phone, city, state, zip)
values ('Steve','Jobs','00112233','6666669','San Cupertino','Calafornia','10001');
