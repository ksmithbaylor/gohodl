package kevin

import (
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

// Rename this to `handleWhatever` for each new type of transaction
func HandlerTemplate(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}
