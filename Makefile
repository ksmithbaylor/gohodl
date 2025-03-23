.PHONY: refresh

refresh:
	SKIP_EXPORT=true go run cmd/ctc/main.go
	sort -r -t, -k1,3 data/txs.csv -o data/txs.csv

ctc:
	SKIP_IDENTIFY=true SKIP_FETCH=true SKIP_ANALYZE=true go run cmd/ctc/main.go
