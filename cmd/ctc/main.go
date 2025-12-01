package main

import (
	"os"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/ctc"
	"github.com/ksmithbaylor/gohodl/internal/generic"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func main() {
	cfg := config.Config

	db := util.NewFileDB("data")
	clients := generic.NewAllNodeClients(cfg.AllNetworks())

	ctc.IdentifyTransactions(db, clients)
	if _, ok := os.LookupEnv("STOP_AFTER_IDENTIFY"); ok {
		return
	}

	txHashes := ctc.FetchTransactions(db, clients)
	ctc.AnalyzeTransactions(db, txHashes)
	ctc.ExportTransactions(db, clients)
}
