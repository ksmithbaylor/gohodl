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
    and txs.method not in (
      "0x110bcd45",
      "0x12514bba",
      "0x12d94235",
      "0x15270ace",
      "0x163e1e61",
      "0x2c10c112",
      "0x327ca788",
      "0x3fe561cf",
      "0x441ff998",
      "0x4ee51a27",
      "0x4f61d102",
      "0x512d7cfd",
      "0x520f3e69",
      "0x588d826a",
      "0x5c45079a",
      "0x62b74da5",
      "0x67243482",
      "0x6c6c9c84",
      "0x6e56cd92",
      "0x729ad39e",
      "0x74a72e41",
      "0x7c8255db",
      "0x7f4d683a",
      "0x82947abe",
      "0x927f59ba",
      "0x9c96eec5",
      "0xa8c6551f",
      "0xb8ae5a2c",
      "0xbd075b84",
      "0xc01ae5d3",
      "0xc204642c",
      "0xc73a2d60",
      "0xd43a632f",
      "0xd57498ea",
      "0xeeb9052f",
      "0xfaf67b43"
    )
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
