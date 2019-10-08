package query

type (
	PublishedReceiptQueryInterface interface {
	}

	PublishedReceiptQuery struct {
		Fields    []string
		TableName string
	}
)

// NewPublishedReceiptQuery returns PublishedQuery instance
func NewPublishedReceiptQuery() *ReceiptQuery {
	return &ReceiptQuery{
		Fields: []string{
			"sender_public_key",
			"recipient_public_key",
			"datum_type",
			"datum_hash",
			"reference_block_height",
			"reference_block_hash",
			"rmr_linked",
			"recipient_signature",
			"intermediate_hashes",
			"block_height",
			"receipt_index",
		},
		TableName: "published_receipt",
	}
}

func (prq *PublishedReceiptQuery) getTableName() string {
	return prq.TableName
}
