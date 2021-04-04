package identity

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/MichaelMure/git-bug-migration/migration3/after/entity"
	"github.com/MichaelMure/git-bug-migration/migration3/after/repository"
)

// Fetch retrieve updates from a remote
// This does not change the local identities state
func Fetch(repo repository.Repo, remote string) (string, error) {
	return repo.FetchRefs(remote, "identities")
}

// Push update a remote with the local changes
func Push(repo repository.Repo, remote string) (string, error) {
	return repo.PushRefs(remote, "identities")
}

// Pull will do a Fetch + MergeAll
// This function will return an error if a merge fail
func Pull(repo repository.ClockedRepo, remote string) error {
	_, err := Fetch(repo, remote)
	if err != nil {
		return err
	}

	for merge := range MergeAll(repo, remote) {
		if merge.Err != nil {
			return merge.Err
		}
		if merge.Status == entity.MergeStatusInvalid {
			return errors.Errorf("merge failure: %s", merge.Reason)
		}
	}

	return nil
}

// MergeAll will merge all the available remote identity
func MergeAll(repo repository.ClockedRepo, remote string) <-chan entity.MergeResult {
	out := make(chan entity.MergeResult)

	go func() {
		defer close(out)

		remoteRefSpec := fmt.Sprintf(identityRemoteRefPattern, remote)
		remoteRefs, err := repo.ListRefs(remoteRefSpec)

		if err != nil {
			out <- entity.MergeResult{Err: err}
			return
		}

		for _, remoteRef := range remoteRefs {
			refSplit := strings.Split(remoteRef, "/")
			id := entity.Id(refSplit[len(refSplit)-1])

			if err := id.Validate(); err != nil {
				out <- entity.NewMergeInvalidStatus(id, errors.Wrap(err, "invalid ref").Error())
				continue
			}

			remoteIdentity, err := read(repo, remoteRef)

			if err != nil {
				out <- entity.NewMergeInvalidStatus(id, errors.Wrap(err, "remote identity is not readable").Error())
				continue
			}

			// Check for error in remote data
			if err := remoteIdentity.Validate(); err != nil {
				out <- entity.NewMergeInvalidStatus(id, errors.Wrap(err, "remote identity is invalid").Error())
				continue
			}

			localRef := identityRefPattern + remoteIdentity.Id().String()
			localExist, err := repo.RefExist(localRef)

			if err != nil {
				out <- entity.NewMergeError(err, id)
				continue
			}

			// the identity is not local yet, simply create the reference
			if !localExist {
				err := repo.CopyRef(remoteRef, localRef)

				if err != nil {
					out <- entity.NewMergeError(err, id)
					return
				}

				out <- entity.NewMergeNewStatus(id, remoteIdentity)
				continue
			}

			localIdentity, err := read(repo, localRef)

			if err != nil {
				out <- entity.NewMergeError(errors.Wrap(err, "local identity is not readable"), id)
				return
			}

			updated, err := localIdentity.Merge(repo, remoteIdentity)

			if err != nil {
				out <- entity.NewMergeInvalidStatus(id, errors.Wrap(err, "merge failed").Error())
				return
			}

			if updated {
				out <- entity.NewMergeUpdatedStatus(id, localIdentity)
			} else {
				out <- entity.NewMergeNothingStatus(id)
			}
		}
	}()

	return out
}
