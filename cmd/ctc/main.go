package main

import (
	"github.com/ksmithbaylor/gohodl/internal/ctc"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

func main() {
	db := util.NewFileDB("data")

	ctc.IdentifyTransactions(db)
}
