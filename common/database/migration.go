package database

import (
	"fmt"

	"github.com/zoobc/zoobc-core/common/query"
)

/*
Migration is struct that included:
	- Init() initialization
	- Apply() run migrations
migration should be has `query.Executor` interface
*/
type Migration struct {
	CurrentVersion *int
	Versions       []string
	Query          *query.Executor
}

/*
Init function must be call at the first time before call `Apply()`.
That just for make sure no error that caused by `query.Executor` not `nil`
and initialize versions
*/
func (m *Migration) Init(query *query.Executor) error {

	if query != nil {
		rows, _ := query.ExecuteSelect("SELECT version FROM migration;")
		if rows != nil {
			var version int
			_ = rows.Scan(&version)
			m.CurrentVersion = &version
		}

		m.Query = query
		m.Versions = []string{
			`CREATE TABLE IF NOT EXISTS "migration" (
				"version" INTEGER DEFAULT 0 NOT NULL,
				"created_date" TIMESTAMP NOT NULL
			);`,
		}
		return nil
	}
	return fmt.Errorf("make sure have add query.Executor")

}

/*
Apply for applying migrations that had initialize on `Init()`.
And this will create migration table included version of migration
*/
func (m *Migration) Apply() error {

	var (
		migrations = m.Versions
	)

	if m.CurrentVersion != nil {
		migrations = m.Versions[*m.CurrentVersion:]
	}

	for version, query := range migrations {
		queries := []string{
			query,
		}
		if m.CurrentVersion != nil {
			queries = append(queries, fmt.Sprintf(`
				UPDATE "migration"
				SET "version" = %d, "created_date" = datetime('now');
			`, *m.CurrentVersion))
		} else {
			queries = append(queries, `
				INSERT INTO "migration" (
					"version",
					"created_date"
				)
				VALUES (
					0,
					datetime('now')
				);
			`)
		}
		_, err := m.Query.ExecuteTransactions(queries)
		m.CurrentVersion = &version
		if err != nil {
			return err
		}
	}
	return nil
}
