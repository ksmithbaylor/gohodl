.mode csv
.import ./data/txs.csv txs

.mode table

create temp view unique_networks as
  select
    network,
    count(*) as how_many
  from txs
  group by network
  order by how_many desc;

select * from by_network;

create temp view unique_methods as
  select
    method,
    count(*) as how_many
  from txs
  group by method
  order by how_many desc;

select count(*) as 'unique methods' from unique_methods;
select
  method,
  how_many,
  sum(how_many) over (order by how_many desc rows unbounded preceding) as cumulative
from unique_methods
limit 20;

create temp view unique_destinations as
  select
    network,
    "to",
    count(*) as how_many
  from txs
  group by network, "to"
  order by how_many desc;

select count(*) as 'unique destinations' from unique_destinations;
select
  network,
  "to",
  how_many,
  sum(how_many) over (order by how_many desc rows unbounded preceding) as cumulative
from unique_destinations
limit 20;

create temp view unique_calls as
  select
    network,
    "to",
    method,
    count(*) as how_many
  from txs
  group by network, "to", method
  order by how_many desc;

select count(*) as 'unique calls' from unique_calls;
select
  network,
  "to",
  method,
  how_many,
  sum(how_many) over (order by how_many desc rows unbounded preceding) as cumulative
from unique_calls
limit 50;
