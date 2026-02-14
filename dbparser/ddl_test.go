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
