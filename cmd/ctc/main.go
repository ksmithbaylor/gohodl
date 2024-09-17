package main

import (
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
	ctc.FetchTransactions(db, clients)
	ctc.AnalyzeTransactions(db)
}
