package kevin

import (
	"fmt"
	"time"

	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func printHeader(bundle handlers.TransactionBundle) {
	t := time.Unix(int64(bundle.Block.Time), 0).UTC().Format("2006-01-02 15:04:05")

	fmt.Printf("----------------- %s: %-12s/ %s (%s -> %s)\n",
		t,
		bundle.Info.Network,
		bundle.Info.Hash,
		bundle.Info.From,
		bundle.Info.To,
	)
}
