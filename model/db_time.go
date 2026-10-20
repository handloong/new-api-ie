package model

import (
	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// GetDBTimestamp returns a UNIX timestamp from database time.
// Falls back to application time on error.
func GetDBTimestamp() int64 {
	return dbTimestamp(nil)
}

// getDBTimestampTx behaves like GetDBTimestamp but runs the query on the given
// transaction when one is supplied. Reading the clock through a separate pooled
// connection inside an open transaction is both wasteful and, with a single
// connection pool, a deadlock.
func getDBTimestampTx(tx *gorm.DB) int64 {
	return dbTimestamp(tx)
}

func dbTimestamp(tx *gorm.DB) int64 {
	db := tx
	if db == nil {
		db = DB
	}
	var ts int64
	var err error
	switch {
	case common.UsingMainDatabase(common.DatabaseTypePostgreSQL):
		err = db.Raw("SELECT EXTRACT(EPOCH FROM NOW())::bigint").Scan(&ts).Error
	case common.UsingMainDatabase(common.DatabaseTypeSQLite):
		err = db.Raw("SELECT strftime('%s','now')").Scan(&ts).Error
	default:
		err = db.Raw("SELECT UNIX_TIMESTAMP()").Scan(&ts).Error
	}
	if err != nil || ts <= 0 {
		return common.GetTimestamp()
	}
	return ts
}
