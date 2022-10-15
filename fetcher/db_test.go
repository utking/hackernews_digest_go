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

func TestRepositoryInit(t *testing.T) {
	repo := DataRepository{dbConfig: Database{Driver: "sqlite3", Database: ":memory:"}}

	repo.Init()
	defer repo.Close()

	items, _ := repo.GetExistingIDs()

	if len(items) > 0 {
		t.Errorf("Expected 0 IDs in the repository, %d exist", len(items))
	}
}

func TestRepositoryUpdateItems(t *testing.T) {
	repo := DataRepository{dbConfig: Database{Driver: "sqlite3", Database: ":memory:"}}

	err := repo.Init()

	if err != nil {
		t.Errorf("Error while preparing a test database in memory, %v", err)
	}

	defer repo.Close()

	items, _ := repo.GetExistingIDs()

	if len(items) > 0 {
		t.Errorf("Expected 0 IDs in the repository, %d exist", len(items))
	}

	repo.UpdateItems([]DigestItem{
		{id: 111, newsTitle: "Some Ititem", newsUrl: "http://localhost", createdAt: 123456789},
	})

	items, _ = repo.GetExistingIDs()

	if len(items) != 1 {
		t.Errorf("Expected 1 ID in the repository, %d exist", len(items))
	}

	for id := range items {
		if id != 111 {
			t.Errorf("Expected ID to be %d, %d found", 111, id)
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
