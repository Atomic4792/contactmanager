\c template1
drop database if exists advancedid;

create database advancedid;
GRANT CONNECT ON DATABASE advancedid TO advanced;
GRANT ALL PRIVILEGES ON DATABASE advancedid TO advanced;
ALTER USER advanced WITH PASSWORD 'set!Application:Password'; -- This is in the config, should be modified before imported then deleted

\c advancedid;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;
CREATE EXTENSION postgis;


-- need to verify apt-get install postgresql-10-postgis
-- CREATE EXTENSION IF NOT EXISTS "postgis" WITH SCHEMA public;

drop type if exists type_of_user;
drop type if exists pay_types;
create type type_of_user as enum  ('ENTERPRISE','CONSUMER');
create type pay_types as enum  ('WIRE','ACH', 'CHECK' ,'BANK TRANSFER');

create table admins ( -- admin users
  id uuid primary key default uuid_generate_v4(),
  email text,
  name text,
  phone bigint,
  password text,
  enabled bool default true,
  created timestamp default now()
);

insert into admins (email,name,phone,password) values ('jpastore79@gmail.com','Jon Pastore','5614656666','abc123');
insert into admins (email,name,phone,password) values ('bharat@bramsoft.com','Bharat Ramaswamy','3107447928','abc123');

create table audit_admins(
  id uuid primary key default uuid_generate_v4(),
  admin_id uuid references admins(id),
  created timestamp default now(),
  action text
);

create table account_types (
  id uuid primary key default uuid_generate_v4(),
  admin_id uuid references admins(id),
  created timestamp default now(),
  max_users int4 default 1,
  monthly_price numeric(6,2),
  annual_price numeric(9,2),
  corp_plan bool default false,
  name text,
  active bool default true
);

insert into account_types (name, corp_plan, max_users, monthly_price, annual_price)
values
('Personal',false,1,2,10),
('Family',false,6,1,8),
('Small',true,10,1.5,9),
('Medium',true,100,1.25,8),
('Large',true,1000,1,7)
;

create table resellers ( -- this is the app company to all for white label
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  created_by uuid references admins(id),
  approved bool,
  approved_by uuid references admins(id),
  name text,
  phone bigint,
  primary_name text,
  primary_phone bigint,
  email text,
  password text,
  verification_string uuid,
  verified timestamp,
  address1 text,
  address2 text,
  city text,
  state text,
  zip text,
  rate_per_user numeric(6,2), -- default to set companies to
  commission_rate numeric(3,2) default .5, -- amount to split revenue 50/50 = .5
  apikey uuid -- apikey for AI, allows creating sub keys for companies
);

create table audit_resellers(
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  reseller_id uuid references resellers(id),
  action text
);

create table companies ( -- this is the app company to all for white label
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  reseller_id uuid references resellers(id),
  name text,
  phone bigint,
  primary_contact text,
  primary_contact_phone bigint,
  primary_email text,
  enabled bool default true,
  address1 text,
  address2 text,
  city text,
  state text,
  zip text,
  apikey uuid,
  num_users int4, -- adding new users bills for the new user during the billing cycle as a whole, user included in renewal
  start_date date default date(now()),
  last_invoice date,
  expiration date,
  rate_per_user numeric(6,2)
);

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

-- invoices should be settled
create table reseller_invoices(
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  reseller_id uuid references resellers(id),
  invoice_month text, -- YYYY-MM
  amount_billed numeric(14,2),
  amount_received numeric(14,2),
  due date default date(now()+interval'7 days')
);

create table reseller_payments(
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  entered_by uuid references admins(id),
  reseller_id uuid references resellers(id),
  invoice_id uuid references reseller_invoices(id),
  payment_method pay_types
);
create index reseller_payments_resellers_idx on reseller_payments(reseller_id);
create index reseller_payments_date_idx on reseller_payments(created);

-- invoices should be created 30 days from last_invoice
-- if they disable, then enable again, the start date should be the day enabled and last_invoice would be nulled
create table company_invoices(
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  reseller_id uuid references resellers(id),
  company_id uuid references companies(id),
  amount_billed numeric(14,2),
  amount_received numeric(14,2),
  due date default date(now()+interval'3 days')
);

create table company_payments(
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  company_id uuid references companies(id),
  invoice_id uuid references company_invoices(id),
  cc_amount numeric(12,2),
  cc_auth text -- authorization code for a successful transaction
);

-- maybe merge above?
create table failed_company_payments (
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  company_id uuid references companies(id)

);

create sequence tos_ver_seq;
create table tos(
  id uuid primary key default uuid_generate_v4(),
  created timestamp default now(),
  created_by uuid references admins(id),
  version int4 default nextval('tos_ver_seq'),
  tos text
);

insert into tos (created_by, version, tos)
select id,1,encode('This is our first ToS', 'base64') from admins limit 1;

drop table if exists user_upload_ingest;
create table user_upload_ingest (
  id uuid primary key default uuid_generate_v4(),
    company_id uuid,
    user_list jsonb,
    filename text,
    processed bool default false,
    ingested timestamp default now()
);

drop table if exists user_upload_staging;
create table user_upload_staging (
  id uuid primary key default uuid_generate_v4(),
  company_id uuid,
  name text,
  phone bigint,
  email text,
  tz_offset text,
  status text,
  imported timestamp default now()
);

create table users ( -- app users
  id uuid primary key default uuid_generate_v4(),
  enabled bool default false,
  company_id uuid references companies(id),
  parent_id uuid, -- sets the parent of company or family in future ver
  user_type type_of_user, -- enterprise or consumer
  tz_offset text,
  name text,
  tos_ver int4,
  tos_updated timestamp,
  tos_ip inet,
  email text,
  password text,
  email_verification_sent timestamp,
  email_verification uuid default uuid_generate_v4(),
  email_verified timestamp,
  verified_ip inet,
  phone bigint,
  npanxx bigint,
  block_id text,
  ocn text,
  lata int4,
  rc_lata int4,
  lata_updated timestamp,
  attributes jsonb,
  os text,
  os_ver text,
  sms_verification_sent timestamp,
  sms_verification_code int4,
  sms_verification_code_generated timestamp,
  sms_verified timestamp,
  created timestamp default now(),
  updated timestamp default now(),
  legal_adult bool default false
);
create index users_phone_idx on users(phone);

create table users_history (
  id uuid,
  enabled bool,
  company_id uuid,
  parent_id uuid,
  user_type type_of_user,
  tz_offset text,
  name text,
  tos_ver int4,
  tos_updated timestamp,
  tos_ip inet,
  email text,
  password text,
  email_verification_sent timestamp,
  email_verification uuid,
  email_verified timestamp,
  verified_ip inet,
  phone bigint,
  npanxx bigint,
  block_id text,
  ocn text,
  lata int4,
  rc_lata int4,
  lata_updated timestamp,
  attributes jsonb,
  os text,
  os_ver text,
  sms_verification_sent timestamp,
  sms_verification_code int4,
  sms_verification_code_generated timestamp,
  sms_verified timestamp,
  created timestamp,
  updated timestamp,
  legal_adult bool
);
create index users_history_phone_idx on users(phone);

create table user_locations (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id),
  sampled timestamp default now(),
  lat float,
  long float,
  where_is geography
);
create index user_locations_user_id_sampled_idx on user_locations(user_id, sampled);

create table sessions (
  id uuid primary key default uuid_generate_v4(),
  expires timestamp default now() + interval'7 days',
  phone bigint,
  device_id uuid,
  user_id uuid
);
create index sessions_phone_device_id_idx on sessions(phone, device_id);

create table sessions_history (
  id uuid primary key default uuid_generate_v4(),
  expires timestamp default now() + interval'7 days',
  phone bigint,
  device_id uuid,
  user_id uuid,
  archived timestamp default now()
);

create index sessions_history_phone_idx on sessions_history(phone);

create table audit_users(
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id),
  user_agent text,
  created timestamp default now(),
  ip inet,
  action text
);

create table user_tos (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id),
  version int4,
  accepted timestamp default now(),
  user_agent text,
  ip inet
);
create index user_tos_user_id_idx on user_tos(user_id);


drop type if exists contact_status ;
create type contact_status as enum  ('WL','BL','EX','NA');

create table user_contacts_history (
  id uuid,
  user_id uuid references users(id),
  phone bigint,
  name text,
  status contact_status,
  created timestamp default now(),
  archived timestamp default now()
);
create index user_contacts_history_user_id_name_idx
    on user_contacts_history(user_id, name);

create table user_contacts (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id),
  phone bigint,
  name text,
  status contact_status,
  created timestamp default now()
);
create index user_contacts_user_id_name_idx on user_contacts(user_id, name);

create table user_schedule (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id),
  day int4, -- 0 for sunday, 6 for sat
  start time,
  duration int4 -- in minutes
);

create table email_verification_staging(
  verified timestamp,
  ip inet,
  user_id uuid,
  phone bigint,
  email_verification uuid
);

create table sms_verification_staging(
  verified timestamp,
  ip inet,
  device_id uuid
);


create table lerg6_staging (
  id uuid primary key default uuid_generate_v4(),
  lerg_date date,
  effective text,
  effective_date date,
  lata int4,
  npanxx int4,
  block_id text,
  ocn text,
  start_lines int4,
  end_lines int4,
  rc_lata int4,
  created timestamp default now()
);

create table lerg6 (
  id uuid primary key default uuid_generate_v4(),
  lerg_date date,
  effective date,
  ocn text,
  lata int4,
  rc_lata int4,
  npanxx bigint,
  block_id text,
  start_lines int4,
  end_lines int4,
  created timestamp default now()
);

create table lerg6_history (
  id uuid,
  lerg_date date,
  effective date,
  ocn text,
  lata int4,
  rc_lata int4,
  npanxx bigint,
  block_id text,
  start_lines int4,
  end_lines int4,
  created timestamp default now()
);

create table user_reports (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id),
  ani bigint,
  ocn text,
  lrn bigint,
  call_ts timestamp,
  callback_response text, -- disconnected, spoofed, automated system
  spoofed bool default false,
  call_type text, -- warranties, prison, charity, insurance, student loan, debt, debt collector
  previous_relationship bool default false,
  optout_provided bool default false,
  tried_optout bool default false,
  created timestamp default now()
);
