package mysql

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

type TestUser struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
}

func TestMysqlDriver(t *testing.T) {
	f, err := os.Create("./testdata/test.log")
	require.Nil(t, err)
	defer f.Close()

	d, err := NewDriver(NewOption(
		WithHost("127.0.0.1"),
		WithPassword("123456"),
		WithUserName("root"),
		WithDbName("testdb"),
		WithLoggerOut(f),
		WithLoggerLevel(logger.Info),
	))
	if err != nil {
		t.Skip(err.Error())
	}

	require.Nil(t, d.AutoMigrate(TestUser{}))
	defer func() {
		assert.Nil(t, d.Exec("drop table test_user;").Error)
	}()
	assert.Nil(t, d.Create(&TestUser{
		Name: "user1",
	}).Error)
}
