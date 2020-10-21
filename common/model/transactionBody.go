package model

// TransactionBodyInterface allowing isTransaction_TransactionBody access from other package
type TransactionBodyInterface = isTransaction_TransactionBody

func (*NodeRegistrationTransactionBody) isTransaction_TransactionBody()       {}
func (*UpdateNodeRegistrationTransactionBody) isTransaction_TransactionBody() {}
func (*RemoveNodeRegistrationTransactionBody) isTransaction_TransactionBody() {}
func (*ClaimNodeRegistrationTransactionBody) isTransaction_TransactionBody()  {}
func (*EmptyTransactionBody) isTransaction_TransactionBody()                  {}
func (*SendMoneyTransactionBody) isTransaction_TransactionBody()              {}
func (*SetupAccountDatasetTransactionBody) isTransaction_TransactionBody()    {}
func (*RemoveAccountDatasetTransactionBody) isTransaction_TransactionBody()   {}
func (*ApprovalEscrowTransactionBody) isTransaction_TransactionBody()         {}
func (*MultiSignatureTransactionBody) isTransaction_TransactionBody()         {}
func (*FeeVoteCommitTransactionBody) isTransaction_TransactionBody()          {}
func (*FeeVoteRevealTransactionBody) isTransaction_TransactionBody()          {}
func (*LiquidPaymentTransactionBody) isTransaction_TransactionBody()          {}
func (*LiquidPaymentStopTransactionBody) isTransaction_TransactionBody()      {}
func (*AtomicTransactionBody) isTransaction_TransactionBody()                 {}
