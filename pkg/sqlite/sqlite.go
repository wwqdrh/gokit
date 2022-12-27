package sqlite

import (
	"errors"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type SqliteDriver struct {
	*gorm.DB

	opt *option
}

func NewSqliteDriver(opt *option) (*SqliteDriver, error) {
	driver := &SqliteDriver{
		opt: opt,
	}
	if err := driver.Connection(); err != nil {
		return nil, err
	}
	return driver, nil
}

func (d *SqliteDriver) Connection() error {
	if d.opt.DbName == "" {
		return errors.New("数据库名字不能为空")
	}

	db, err := gorm.Open(sqlite.Open(d.opt.DbName), &gorm.Config{
		Logger: logger.New(
			log.New(d.opt.loggerOut, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second,       // Slow SQL threshold
				LogLevel:                  d.opt.loggerLevel, // Log level
				IgnoreRecordNotFoundError: true,              // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,             // Disable color
			},
		),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 必须添加这个 在migrate的时候才不会创建外键
	})

	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	d.DB = db
	return nil
}
