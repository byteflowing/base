package db

import (
	"gorm.io/gorm"
	"gorm.io/rawsql"

	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

func New(c *configv1.DbConfig) *gorm.DB {
	var db *gorm.DB
	switch c.DbType {
	case enumv1.DbType_DB_TYPE_MYSQL:
		db = initMySQL(c)
	case enumv1.DbType_DB_TYPE_POSTGRES:
		db = initPostgres(c)
	case enumv1.DbType_DB_TYPE_SQLSERVER:
		db = initSQLServer(c)
	case enumv1.DbType_DB_TYPE_SQLITE:
		db = initSQLite(c)
	default:
		panic("unknown database type:" + c.DbType.String())
	}
	if c.DbType != enumv1.DbType_DB_TYPE_SQLITE && c.Conn != nil {
		sqlDb, err := db.DB()
		if err != nil {
			panic(err)
		}
		sqlDb.SetConnMaxIdleTime(getMaxIdleTime(c.Conn))
		sqlDb.SetConnMaxLifetime(getConnMaxLifetime(c.Conn))
		sqlDb.SetMaxIdleConns(getMaxIdleConnes(c.Conn))
		sqlDb.SetMaxOpenConns(getMaxOpenConnes(c.Conn))
	}
	return db
}

// NewBySQL 通过SQL创建db,主要用于gen生成struct
func NewBySQL(c *configv1.SqlConfig) *gorm.DB {
	conf := rawsql.Config{
		DriverName: getDBType(c.DbType),
		FilePath:   c.FilePath,
		SQL:        c.Sql,
	}
	gormDB, err := gorm.Open(rawsql.New(conf))
	if err != nil {
		panic(err)
	}
	return gormDB
}

func getDBType(t enumv1.DbType) string {
	switch t {
	case enumv1.DbType_DB_TYPE_MYSQL:
		return "mysql"
	case enumv1.DbType_DB_TYPE_POSTGRES:
		return "postgres"
	case enumv1.DbType_DB_TYPE_SQLSERVER:
		return "sqlserver"
	case enumv1.DbType_DB_TYPE_SQLITE:
		return "sqlite"
	default:
		panic("unknown database type:" + t.String())
	}
}
