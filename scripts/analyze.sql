.mode csv
.import ./data/txs.csv txs
.import ./data/ctc.csv ctc

.mode table

create temp view unique_networks as
  select
    network,
    count(*) as how_many
  from txs
  where txs.hash not in (select substr("ID (Optional)", 0, 67) from ctc)
    and txs.timestamp > 1704067199
    and txs.timestamp <= 1735689599
  group by network
  order by how_many desc;

select * from unique_networks;

create temp view unique_methods as
  select
    method,
    count(*) as how_many,
    count(distinct "to") as destinations
  from txs
  where txs.hash not in (select substr("ID (Optional)", 0, 67) from ctc)
    and txs.timestamp > 1704067199
    and txs.timestamp <= 1735689599
    and txs.method not in ("0x9c96eec5", "0x26ededb8", "0x441ff998", "0x729ad39e")
  group by method
  order by how_many desc;

select
  method,
  how_many,
  destinations,
  sum(how_many) over (order by how_many desc, method rows unbounded preceding) as cumulative
from unique_methods
order by how_many, method desc;
select count(*) as 'unique methods' from unique_methods;

-- create temp view unique_destinations as
--   select
--     network,
--     "to",
--     count(*) as how_many,
--     count(distinct method) as methods
--   from txs
--   where txs.hash not in (select substr("ID (Optional)", 0, 67) from ctc)
--     and txs.timestamp > 1704067199
--     and txs.timestamp <= 1735689599
--   group by network, "to"
--   order by how_many desc;
--
-- select count(*) as 'unique destinations' from unique_destinations;
-- select
--   network,
--   "to",
--   how_many,
--   methods,
--   sum(how_many) over (order by how_many desc rows unbounded preceding) as cumulative
-- from unique_destinations
-- limit 20;
--
-- create temp view unique_calls as
--   select
--     network,
--     "to",
--     method,
--     count(*) as how_many
--   from txs
--   where txs.hash not in (select substr("ID (Optional)", 0, 67) from ctc)
--     and txs.timestamp > 1704067199
--     and txs.timestamp <= 1735689599
--   group by network, "to", method
--   order by how_many desc;
--
-- select count(*) as 'unique calls' from unique_calls;
-- select
--   network,
--   "to",
--   method,
--   how_many,
--   sum(how_many) over (order by how_many desc rows unbounded preceding) as cumulative
-- from unique_calls
-- limit 50;
