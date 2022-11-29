package fetcher

import (
	"testing"
)

func TestPrepareRepository(t *testing.T) {
	repo := DataRepository{dbConfig: Database{Driver: "sqlite3", Database: ":memory:"}}

	err := repo.prepareDB()
	defer repo.Close()

	if err != nil {
		t.Errorf("Error while preparing a test database in memory, %v", err)
	}
}

func TestPrepareRepositoryWrongDriver(t *testing.T) {
	repo := DataRepository{dbConfig: Database{Driver: "bad-driver", Database: ":memory:"}}

	err := repo.prepareDB()

	if err == nil {
		t.Errorf("Error must have been thrown while preparing a test database in memory")
	}
}

func TestRepositoryUpdateItems(t *testing.T) {
	repo := DataRepository{dbConfig: Database{Driver: "sqlite3", Database: ":memory:"}}

	err := repo.Init()

	if err != nil {
		t.Errorf("Error while preparing a test database in memory, %v", err)
	}

	defer repo.Close()

	digest := &[]DigestItem{
		{id: 111, newsTitle: "Some Ititem", newsUrl: "http://localhost", createdAt: 123456789},
	}

	if err := repo.UpdateItems(digest); err != nil {
		t.Errorf("Could not update the repository")
	}

	items, _ := repo.GetIDsToPull(&[]int64{112})

	if len(items) != 1 {
		t.Errorf("Expected 1 ID not in the repository, %d exist", len(items))
	}

	for _, id := range items {
		if id != 112 {
			t.Errorf("Expected ID to pull to be %d, %d found", 112, id)
		}
	}
}

func TestVacuumRepository(t *testing.T) {
	repo := DataRepository{dbConfig: Database{Driver: "sqlite3", Database: ":memory:"}}

	err := repo.Init()

	if err != nil {
		t.Errorf("Error while preparing a test database in memory, %v", err)
	}

	defer repo.Close()

	if err := repo.purgeOld(); err != nil {
		t.Errorf("Error while vacuuming the repository, %v", err)
	}
}
