package model

// type TransactionBodyInterface = isTransaction_TransactionBody
type TransactionBodyInterface = isTransaction_TransactionBody

func (*NodeRegistrationTransactionBody) isTransaction_TransactionBody()       {}
func (*UpdateNodeRegistrationTransactionBody) isTransaction_TransactionBody() {}
func (*RemoveNodeRegistrationTransactionBody) isTransaction_TransactionBody() {}
func (*ClaimNodeRegistrationTransactionBody) isTransaction_TransactionBody()  {}
func (*EmptyTransactionBody) isTransaction_TransactionBody()                  {}
func (*SendMoneyTransactionBody) isTransaction_TransactionBody()              {}
func (*SetupAccountDatasetTransactionBody) isTransaction_TransactionBody()    {}
func (*RemoveAccountDatasetTransactionBody) isTransaction_TransactionBody()   {}
