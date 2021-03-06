package migration1

import (
	"fmt"

	"github.com/pkg/errors"

	afterbug "github.com/MichaelMure/git-bug-migration/migration1/after/bug"
	afteridentity "github.com/MichaelMure/git-bug-migration/migration1/after/identity"
	afterrepo "github.com/MichaelMure/git-bug-migration/migration1/after/repository"
)

type Migration1 struct {
	allIdentities []*afteridentity.Identity
}

func (m *Migration1) Description() string {
	return "Convert legacy identities into a complete data structure in git"
}

func (m *Migration1) Run(repoPath string) error {
	repo, err := afterrepo.NewGitRepo(repoPath, nil)
	if err != nil {
		return err
	}

	return m.migrate(repo)
}

func (m *Migration1) migrate(repo afterrepo.ClockedRepo) error {
	err := m.readIdentities(repo)
	if err != nil {
		fmt.Printf("Error while applying migration\n")
		// stop the migration
		return nil
	}

	// Iterating through all the bugs in the repo
	for streamedBug := range afterbug.ReadAllLocal(repo) {
		if streamedBug.Err != nil {
			if streamedBug.Err != afterbug.ErrInvalidFormatVersion {
				fmt.Printf("got error when reading bug, assuming data is already migrated: %q\n", streamedBug.Err)
			} else {
				fmt.Printf("skipping bug, already updated\n")
			}
			continue
		}

		fmt.Printf("%s: ", streamedBug.Bug.Id().Human())

		oldBug := streamedBug.Bug
		newBug, changed, err := m.migrateBug(oldBug, repo)
		if err != nil {
			fmt.Printf("Got error when parsing bug: %q\n", err)
		}

		// If the bug has been changed, remove the old bug and commit the new one
		if changed {
			err = newBug.Commit(repo)
			if err != nil {
				fmt.Printf("Got error when attempting to commit new bug: %q\n", err)
				continue
			}

			err = afterbug.RemoveBug(repo, oldBug.Id())
			if err != nil {
				fmt.Printf("Got error when attempting to remove bug: %q\n", err)
				continue
			}

			fmt.Printf("migrated to %s\n", newBug.Id().Human())
			continue
		}
		fmt.Printf("migration not needed\n")
	}

	return nil
}

func (m *Migration1) readIdentities(repo afterrepo.ClockedRepo) error {
	for streamedIdentity := range afteridentity.ReadAllLocal(repo) {
		if err := streamedIdentity.Err; err != nil {
			if errors.Is(err, afteridentity.ErrIncorrectIdentityFormatVersion) {
				fmt.Print("skipping identity, already updated\n")
				continue
			} else {
				fmt.Printf("Got error when reading identity: %q", streamedIdentity.Err)
				return streamedIdentity.Err
			}
		}
		m.allIdentities = append(m.allIdentities, streamedIdentity.Identity)
	}
	return nil
}

func (m *Migration1) migrateBug(oldBug *afterbug.Bug, repo afterrepo.ClockedRepo) (*afterbug.Bug, bool, error) {
	if oldBug.Packs[0].FormatVersion != 1 {
		return nil, false, nil
	}

	// Making a new bug
	newBug := afterbug.NewBug()

	// Iterating over each operation in the bug
	it := afterbug.NewOperationIterator(oldBug)
	for it.Next() {
		operation := it.Value()
		oldAuthor := operation.GetAuthor()

		// Checking if the author is of the legacy (bare) type
		switch oldAuthor.(type) {
		case *afteridentity.Bare:
			// Search existing identities for any traces of this old identity
			var newAuthor *afteridentity.Identity = nil
			for _, identity := range m.allIdentities {
				if oldAuthor.Name() == identity.Name() {
					newAuthor = identity
				}
			}

			// If no existing identity is found, create a new one
			if newAuthor == nil {
				newAuthor = afteridentity.NewIdentityFull(
					oldAuthor.Name(),
					oldAuthor.Email(),
					oldAuthor.Login(),
					oldAuthor.AvatarUrl(),
				)

				err := newAuthor.Commit(repo)
				if err != nil {
					return nil, false, err
				}
			}

			// Set the author of the operation to the new identity
			operation.SetAuthor(newAuthor)
			newBug.Append(operation)
			continue

		// If the author's identity is a new identity type, its fine. Just append it to the cache
		case *afteridentity.Identity:
			newBug.Append(operation)
			continue

		// This should not be reached
		default:
			return newBug, false, fmt.Errorf("Unknown author type: %T\n", operation.GetAuthor())
		}
	}

	return newBug, true, nil
}
