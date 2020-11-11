package helper

import (
	"database/sql"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
)

// GetSqliteDB to get sql.Db of sqlite based on DB path & name
func GetSqliteDB(dbPath, dbName string) (*sql.DB, error) {
	var sqliteDbInstance = database.NewSqliteDB()
	if err := sqliteDbInstance.InitializeDB(dbPath, dbName); err != nil {
		return nil, err
	}
	sqliteDB, err := sqliteDbInstance.OpenDB(
		dbPath,
		dbName,
		constant.SQLMaxOpenConnetion,
		constant.SQLMaxIdleConnections,
		constant.SQLMaxConnectionLifetime,
	)
	if err != nil {
		return nil, err
	}
	return sqliteDB, nil
}
