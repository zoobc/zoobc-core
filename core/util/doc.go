// Package util (core/util) handle block and transaction common used utility.
// Transaction Bytes is represented as: {TransactionType(2), Timestamp(8), SenderAccountID(32), RecipientAccountID(32),
// Fee(8), TransactionBodyLength(8), TransactionBodyBytes(tbl), Signature(64)}
// tbl: transaction body length
// Block Bytes is represented as: {Version(4), Timestamp(8), NumberOfTransaction(txNumber), TotalAmount(8), TotalFee(8),
// TotalCoinbase(8), PayloadLength(8), PayloadHash(custom), BlocksmithID(32), BlockSeed(custom), PreviousBlockHash(custom),
// Signature(64)}
package util
