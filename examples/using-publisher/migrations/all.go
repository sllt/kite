package migrations

import (
	"github.com/sllt/kite/pkg/kite/migration"
)

func All() map[int64]migration.Migrate {
	return map[int64]migration.Migrate{
		1721801313: createTopics(),
	}
}
