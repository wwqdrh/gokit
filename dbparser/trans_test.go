package parser

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestUserModel struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
}

func (TestUserModel) TableName() string {
	return "test_user"
}

func TestGenCreateTableStatement(t *testing.T) {
	sqliteSQL, err := GenCreateTableStatement(TestUserModel{}, "test_user", "sqlite")
	assert.NoError(t, err)
	fmt.Println(sqliteSQL)

	mysqlSQL, err := GenCreateTableStatement(TestUserModel{}, "test_user", "mysql")
	assert.NoError(t, err)
	fmt.Println(mysqlSQL)
}
