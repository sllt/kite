package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sllt/kite/pkg/kite/migration"
)

func TestAll(t *testing.T) {
	// Get the map of migrations
	allMigrations := All()

	expected := map[int64]migration.Migrate{
		1722507126: createTableEmployee(),
		1722507180: addEmployeeInRedis(),
	}

	// Check if the length of the maps match
	assert.Equal(t, len(expected), len(allMigrations), "TestAll Failed!")
}
