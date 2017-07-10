package storage

import (
	"testing"
	"github.com/bigblind/marvin/storage"
	"github.com/bigblind/marvin/actions/domain"
	"time"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func newTestChore(id string) domain.Chore {
	return domain.Chore{
		ID: id,
		Name: "Test Chore",
		Created: time.Now(),

		Actions: []domain.ActionInstance{
			domain.ActionInstance{
				ID: "1",
				ActionProvider: "testProvider",
				Action: "testAction",
			},
		},
	}
}

func TestSaveAndGetChore(t *testing.T) {
	storage.WithTestDB(t, func(dbs storage.Store) {
		s := NewChoreStore(dbs)
		c := newTestChore("chore_1")
		err := s.SaveChore("account_1", c)
		require.NoError(t, err)

		c2, err := s.GetChore("account_1", "chore_1")
		require.NoError(t, err)

		require.Equal(t, c, c2, "The saved and retrieved chores must be equal.")
	})
}

func TestGetAccountChores(t *testing.T) {
	storage.WithTestDB(t, func(dbs storage.Store) {
		s := NewChoreStore(dbs)
		c1 := newTestChore("chore_1")
		c2 := newTestChore("chore_2")
		c3 := newTestChore("chore_3")
		err := s.SaveChore("account_1", c1)
		err = s.SaveChore("account_1", c2)
		err = s.SaveChore("account_1", c3)
		require.NoError(t, err)


		cs, err := s.GetAccountChores("account_1")
		require.NoError(t, err)

		// Using assert here rather than require, because it doesn't stop after the first fialure,
		// So we know for all 3 whether they're in there.
		assert.Contains(t, cs, c1)
		assert.Contains(t, cs, c2)
		assert.Contains(t, cs, c3)
	})
}


