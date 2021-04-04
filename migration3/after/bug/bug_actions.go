package bug

import (
	"github.com/pkg/errors"

	"github.com/MichaelMure/git-bug-migration/migration3/after/entity"
	"github.com/MichaelMure/git-bug-migration/migration3/after/entity/dag"
	"github.com/MichaelMure/git-bug-migration/migration3/after/identity"
	"github.com/MichaelMure/git-bug-migration/migration3/after/repository"
)

// Fetch retrieve updates from a remote
// This does not change the local bugs state
func Fetch(repo repository.Repo, remote string) (string, error) {
	return dag.Fetch(def, repo, remote)
}

// Push update a remote with the local changes
func Push(repo repository.Repo, remote string) (string, error) {
	return dag.Push(def, repo, remote)
}

// Pull will do a Fetch + MergeAll
// This function will return an error if a merge fail
func Pull(repo repository.ClockedRepo, remote string, author identity.Interface) error {
	_, err := Fetch(repo, remote)
	if err != nil {
		return err
	}

	for merge := range MergeAll(repo, remote, author) {
		if merge.Err != nil {
			return merge.Err
		}
		if merge.Status == entity.MergeStatusInvalid {
			return errors.Errorf("merge failure: %s", merge.Reason)
		}
	}

	return nil
}

// MergeAll will merge all the available remote bug
// Note: an author is necessary for the case where a merge commit is created, as this commit will
// have an author and may be signed if a signing key is available.
func MergeAll(repo repository.ClockedRepo, remote string, author identity.Interface) <-chan entity.MergeResult {
	// no caching for the merge, we load everything from git even if that means multiple
	// copy of the same entity in memory. The cache layer will intercept the results to
	// invalidate entities if necessary.
	identityResolver := identity.NewSimpleResolver(repo)

	out := make(chan entity.MergeResult)

	go func() {
		defer close(out)

		results := dag.MergeAll(def, repo, identityResolver, remote, author)

		// wrap the dag.Entity into a complete Bug
		for result := range results {
			result := result
			if result.Entity != nil {
				result.Entity = &Bug{
					Entity: result.Entity.(*dag.Entity),
				}
			}
			out <- result
		}
	}()

	return out
}

// RemoveBug will remove a local bug from its entity.Id
func RemoveBug(repo repository.ClockedRepo, id entity.Id) error {
	return dag.Remove(def, repo, id)
}
