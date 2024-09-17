package private

// import (
//   "github.com/ksmithbaylor/gohodl/internal/evm"
// )
//
// func init() {
//   Implementation = example(struct{}{})
// }
//
// type example struct{}
//
// func (e example) HandleTransaction(info *evm.TxInfo, readTransaction transactionReader, export ctcWriter) error {
//   readAndThen := func(handle transactionHandler) error {
//     tx, receipt, err := readTransaction(info.Network, info.Hash)
//     if err != nil {
//       return err
//     }
//     return handle(transactionBundle{info, tx, receipt}, export)
//   }
//
//   var err error
//
//   switch info.Method {
//   case "0xa9059cbb": // ERC-20 transfer, as an example
//     err = readAndThen(handleTransfer)
//   }
//
//   return err
// }
//
// ////////////////////////////////////////////////////////////////////////////////
// // Transaction handler functions
//
// func handleTransfer(bundle transactionBundle, export ctcWriter) error {
//   // Do whatever is needed to determine what CTC csv row should result from this
//   // transaction, then `return export([]string{...})` with that data
//   return nil
// }
