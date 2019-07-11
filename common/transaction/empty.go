package transaction

type TXEmpty struct {
}

func (tx *TXEmpty) Apply() error {
	return nil
}
func (tx *TXEmpty) Unconfirmed() error {
	return nil
}
func (tx *TXEmpty) Validate() error {
	return nil
}
