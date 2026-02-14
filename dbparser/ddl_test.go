package parser

import (
	"errors"
	"testing"
)

var (
	testCreateTable = `
create table if not exists user(
	id int primary key auto_increment,
	username varchar(20) not null unique,
	password varchar(100) not null,
	phone varchar(11) not null unique
	);
`

	testCreateTable2 = `
create table user(
	id int primary key auto_increment,
	username varchar not null unique,
	password varchar(100) not null,
	phone varchar(11) not null unique
	);
`

	testCreateTable3 = `
CREATE TABLE TEST_TABLE_2 (ID int primary key, VAL text)
`
)

func TestParserCreateSchema(t *testing.T) {
	res, err := NewParser(NewLexer(testCreateTable).Tokenize()).Parse()
	if err != nil {
		t.Error(err)
	}
	if len(res) != 1 {
		t.Error(errors.New("数据表个数不对"))
	}
	if res[0].Name != "user" {
		t.Error()
	}
	if len(res[0].Columns) != 4 {
		t.Error()
	}
	if res[0].Columns[1].Type != "VARCHAR(20)" {
		t.Error()
	}

	res, err = NewParser(NewLexer(testCreateTable2).Tokenize()).Parse()
	if err != nil {
		t.Error(err)
	}
	if res[0].Name != "user" {
		t.Error()
	}
	if len(res[0].Columns) != 4 {
		t.Error()
	}
	if res[0].Columns[1].Type != "varchar" {
		t.Error()
	}

	res, err = NewParser(NewLexer(testCreateTable3).Tokenize()).Parse()
	if err != nil {
		t.Error(err)
	}
	if len(res[0].Columns) != 2 {
		t.Error()
	}
}

func TestGetTableFromSQL(t *testing.T) {
	// Test case 1: Single CREATE TABLE statement
	singleSQL := `
	CREATE TABLE IF NOT EXISTS user (
		id INT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(20) NOT NULL UNIQUE,
		password VARCHAR(100) NOT NULL,
		phone VARCHAR(11) NOT NULL UNIQUE
	);
	`

	// Test extracting user table from single statement
	userTable, err := GetTableFromSQL(singleSQL, "user")
	if err != nil {
		t.Error(err)
	}
	if userTable.Name != "user" {
		t.Error()
	}
	if len(userTable.Columns) != 4 {
		t.Error()
	}
	if userTable.PkName != "id" {
		t.Error()
	}

	// Test case 2: SQLite3 syntax with table-level PRIMARY KEY
	sqliteSQL := `
	CREATE TABLE sqlite_users (
		id INTEGER,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		PRIMARY KEY (id)
	);

	-- 插入默认设备扩展信息数据
INSERT OR IGNORE INTO device_extra (device_id, key, value) VALUES
('device-003', 'storageCapacity', '4TB'),
('device-003', 'usedStorage', '2.5TB'),
('device-004', 'doorStatus', 'closed'),
('device-006', 'doorStatus', 'closed'),
('device-008', 'storageCapacity', '8TB'),
('device-008', 'usedStorage', '6TB');
	`

	sqliteTable, err := GetTableFromSQL(sqliteSQL, "sqlite_users")
	if err != nil {
		t.Error(err)
	}
	if sqliteTable.Name != "sqlite_users" {
		t.Error()
	}
	if len(sqliteTable.Columns) != 3 {
		t.Error()
	}
	if sqliteTable.PkName != "id" {
		t.Error()
	}

	// Test case 3: Case-insensitive table name matching
	caseSQL := `
	CREATE TABLE TEST_TABLE (
		id INT PRIMARY KEY,
		value VARCHAR(50)
	);
	`

	caseTable, err := GetTableFromSQL(caseSQL, "test_table")
	if err != nil {
		t.Error(err)
	}
	if caseTable.Name != "TEST_TABLE" {
		t.Error()
	}
}
