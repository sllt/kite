package migrations

import (
	"github.com/sllt/kite/pkg/kite/migration"
)

func All() map[int64]migration.Migrate {
	return map[int64]migration.Migrate{
		1722507126: createTableEmployee(),
		1722507180: addEmployeeInRedis(),
	}
}
