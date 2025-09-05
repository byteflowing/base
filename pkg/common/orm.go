package common

import (
	"gorm.io/gorm"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/orm"
)

func NewDb(config *configv1.Db) *gorm.DB {
	switch config.DbType {
	case "mysql":
		if config.Mysql == nil {
			panic("mysql config is nil")
		}
	case "postgres":
		if config.Postgres == nil {
			panic("postgres config is nil")
		}
	case "sqlserver":
		if config.Sqlserver == nil {
			panic("sqlite config is nil")
		}
	case "sqlite":
		if config.Sqlite == nil {
			panic("sqlite config is nil")
		}
	default:
		panic("unknown database type: " + config.DbType)
	}
	return orm.New(&orm.Config{
		DbType:    config.DbType,
		Log:       convertDbLogConfig(config.Log),
		Conn:      convertConnConfig(config.Conn),
		MySQL:     convertMySQLConfig(config.Mysql),
		Postgres:  convertPostgresConfig(config.Postgres),
		SQLServer: convertSQLServerConfig(config.Sqlserver),
		SQLite:    convertSQLiteConfig(config.Sqlite),
	})
}

func convertRotationConfig(config *configv1.LogRotation) *orm.LogRotationConfig {
	if config == nil {
		return nil
	}
	return &orm.LogRotationConfig{
		LogFile:    config.LogFile,
		MaxSize:    int(config.MaxSize),
		MaxAge:     int(config.MaxAge),
		MaxBackups: int(config.MaxBackups),
		Compress:   config.Compress,
		LocalTime:  config.LocalTime,
	}
}

func convertDbLogConfig(config *configv1.DbLog) *orm.LogConfig {
	if config == nil {
		return nil
	}
	return &orm.LogConfig{
		SlowThreshold:             uint(config.SlowThreshold),
		Out:                       config.Out,
		Colorful:                  config.Colorful,
		IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundErr,
		ParameterizedQueries:      config.ParameterizedQueries,
		Level:                     config.Level,
		LogRotation:               convertRotationConfig(config.Rotation),
	}
}

func convertConnConfig(config *configv1.DbConn) *orm.ConnConfig {
	if config == nil {
		return nil
	}
	return &orm.ConnConfig{
		ConnMaxLifetime: int(config.ConnMaxLifeTime),
		MaxIdleTime:     int(config.MaxIdleTime),
		MaxIdleConnes:   int(config.MaxIdleConnes),
		MaxOpenConnes:   int(config.MaxOpenConnes),
	}
}

func convertMySQLConfig(config *configv1.DbMysql) *orm.MySQLConfig {
	if config == nil {
		return nil
	}
	return &orm.MySQLConfig{
		Host:         config.Host,
		User:         config.User,
		Password:     config.Password,
		DBName:       config.DbName,
		Port:         int(config.Port),
		Charset:      config.Charset,
		Location:     config.Location,
		ConnTimeout:  int(config.ConnTimeout),
		ReadTimeout:  int(config.ReadTimeout),
		WriteTimeout: int(config.WriteTimeout),
	}
}

func convertPostgresConfig(config *configv1.DbPostgres) *orm.PostgresConfig {
	if config == nil {
		return nil
	}
	return &orm.PostgresConfig{
		Host:     config.Host,
		User:     config.User,
		Password: config.Password,
		DBName:   config.DbName,
		SSLMode:  config.SslMode,
		Port:     int(config.Port),
		TimeZone: config.TimeZone,
		Schema:   config.Schema,
	}
}

func convertSQLServerConfig(config *configv1.DbSQLServer) *orm.SQLServerConfig {
	if config == nil {
		return nil
	}
	return &orm.SQLServerConfig{
		Host:     config.Host,
		User:     config.User,
		Password: config.Password,
		DBName:   config.DbName,
		Port:     int(config.Port),
	}
}

func convertSQLiteConfig(config *configv1.DbSQLite) *orm.SQLiteConfig {
	if config == nil {
		return nil
	}
	return &orm.SQLiteConfig{
		DBPath: config.DbPath,
	}
}
