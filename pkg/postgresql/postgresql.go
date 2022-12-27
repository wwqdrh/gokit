package postgresql

import (
	"errors"
	"fmt"
	"io"

	"database/sql"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresDriver struct {
	*gorm.DB

	opt *option
}

func LoggerFactory(w io.Writer, lev logger.LogLevel) logger.Interface {
	return logger.New(
		log.New(w, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  lev,         // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)
}

func NewDriver(opt *option) (*PostgresDriver, error) {
	driver := &PostgresDriver{
		opt: opt,
	}
	if err := driver.Connection(); err != nil {
		return nil, err
	}
	return driver, nil
}

func (d *PostgresDriver) DNS(db bool) string {
	if db {
		return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", d.opt.UserName, d.opt.Password, d.opt.Host, d.opt.DbName)
	} else {
		return fmt.Sprintf("postgres://%s:%s@%s?sslmode=disable", d.opt.UserName, d.opt.Password, d.opt.Host)
	}
}

func (d *PostgresDriver) Connection() error {
	if d.DB != nil {
		return nil
	}

	try := func() (*gorm.DB, error) {
		db, err := gorm.Open(postgres.New(postgres.Config{
			DSN: d.DNS(true), // data source name // auto configure based on currently Postgres version
		}), &gorm.Config{
			Logger: LoggerFactory(d.opt.loggerOut, d.opt.loggerLevel),
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
func (d *PostgresDriver) CreateDatabase(name string) error {
	db, err := sql.Open("postgres", d.DNS(false))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s" WITH OWNER "postgres" ENCODING 'UTF8';`, name))
	if err != nil {
		return err
	}

	return nil
}

func (d *PostgresDriver) CreateTrigger(sql string, triggerName string) error {
	triggers := d.ListTrigger()
	if _, ok := triggers[triggerName]; ok {
		// return errors.New("触发器已经存在")
		return nil
	}

	if err := d.DB.Exec(sql).Error; err != nil {
		return err
	}
	return nil
}

func (d *PostgresDriver) CreateEventTrigger(sql, triggerName string) error {
	triggers := d.ListEventTrigger()
	if _, ok := triggers[triggerName]; ok {
		// return errors.New("触发器已经存在")
		return nil
	}
	if err := d.DB.Exec(sql).Error; err != nil {
		return err
	}
	return nil
}

func (d *PostgresDriver) DeleteTrigger(triggerName string) error {
	triggers := d.ListTrigger()
	if _, ok := triggers[triggerName]; !ok {
		return errors.New("触发器不存在")
	}

	if err := d.DB.Exec("delete from pg_trigger where tgname = ?", triggerName).Error; err != nil {
		return err
	}
	return nil
}

func (d *PostgresDriver) ListTrigger() map[string]bool {
	var res map[string]bool = map[string]bool{}
	var triggerName string

	rows, err := d.DB.Raw("select tgname from pg_trigger").Rows()
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&triggerName)
		res[triggerName] = true
	}

	return res
}

func (d *PostgresDriver) ListEventTrigger() map[string]bool {
	// select * from pg_event_trigger;
	var res map[string]bool = map[string]bool{}
	var triggerName string

	rows, err := d.DB.Raw("select evtname from pg_event_trigger").Rows()
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&triggerName)
		res[triggerName] = true
	}

	return res
}
