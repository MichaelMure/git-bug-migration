package bug

import "github.com/MichaelMure/git-bug-migration/migration1/before_test/util/git"

var _ Operation = &NoOpOperation{}

// NoOpOperation is an operation that does not change the bug state. It can
// however be used to store arbitrary metadata in the bug history, for example
// to support a bridge feature
type NoOpOperation struct {
	OpBase
}

func (op *NoOpOperation) base() *OpBase {
	return &op.OpBase
}

func (op *NoOpOperation) Hash() (git.Hash, error) {
	return hashOperation(op)
}

func (op *NoOpOperation) Apply(snapshot *Snapshot) {
	// Nothing to do
}

func (op *NoOpOperation) Validate() error {
	return opBaseValidate(op, NoOpOp)
}

func NewNoOpOp(author Person, unixTime int64) *NoOpOperation {
	return &NoOpOperation{
		OpBase: newOpBase(NoOpOp, author, unixTime),
	}
}

// Convenience function to apply the operation
func NoOp(b Interface, author Person, unixTime int64, metadata map[string]string) (*NoOpOperation, error) {
	op := NewNoOpOp(author, unixTime)

	for key, value := range metadata {
		op.SetMetadata(key, value)
	}

	if err := op.Validate(); err != nil {
		return nil, err
	}
	b.Append(op)
	return op, nil
}
