package bug

import (
	"errors"

	"github.com/MichaelMure/git-bug-migration/migration3/after/entity"
)

var ErrBugNotExist = errors.New("bug doesn't exist")

func NewErrMultipleMatchBug(matching []entity.Id) *entity.ErrMultipleMatch {
	return entity.NewErrMultipleMatch("bug", matching)
}

func NewErrMultipleMatchOp(matching []entity.Id) *entity.ErrMultipleMatch {
	return entity.NewErrMultipleMatch("operation", matching)
}
