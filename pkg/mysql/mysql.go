package mysql

import (
	"fmt"

	"database/sql"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type MysqlDriver struct {
	*gorm.DB

	opt *option
}

func NewDriver(opt *option) (*MysqlDriver, error) {
	driver := &MysqlDriver{
		opt: opt,
	}
	if err := driver.Connection(); err != nil {
		return nil, err
	}
	return driver, nil
}

func (d *MysqlDriver) DNS(db bool) string {
	if db {
		return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", d.opt.UserName, d.opt.Password, d.opt.Host, d.opt.DbName)
	} else {
		return fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8mb4&parseTime=True&loc=Local", d.opt.UserName, d.opt.Password, d.opt.Host)
	}
}

func (d *MysqlDriver) Connection() error {
	if d.DB != nil {
		return nil
	}

	try := func() (*gorm.DB, error) {
		db, err := gorm.Open(mysql.Open(d.DNS(true)), &gorm.Config{
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
			return nil, err
		}
		return db, nil
	}

	if db, err := try(); err != nil {
		if err := d.CreateDatabase(d.opt.DbName); err != nil {
			return err
		}
		if db, err := try(); err != nil {
			return err
		} else {
			d.DB = db
		}
	} else {
		d.DB = db
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(d.opt.maxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(d.opt.maxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(d.opt.connMaxLifetime)

	return nil
}

// CreateDatabase 创建新的数据库 指定名字，指定charset 引擎 之类
func (d *MysqlDriver) CreateDatabase(name string) error {
	db, err := sql.Open("mysql", d.DNS(false))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE %s DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_general_ci;`, name))
	if err != nil {
		return err
	}

	return nil
}
