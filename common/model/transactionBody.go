package model

// type TransactionBodyInterface = isTransaction_TransactionBody
type TransactionBodyInterface interface {
	isTxBody()
}

func (*NodeRegistrationTransactionBody) isTxBody()       {}
func (*UpdateNodeRegistrationTransactionBody) isTxBody() {}
func (*EmptyTransactionBody) isTxBody()                  {}
func (*SendMoneyTransactionBody) isTxBody()              {}
