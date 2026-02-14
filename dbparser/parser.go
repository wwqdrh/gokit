package dbparser

import (
	"errors"
	"strings"
)

type SQLType int

const (
	DDL SQLType = iota
	DQL
	Unknown
)

func GetSqlType(sql string) SQLType {
	sql = strings.TrimSpace(sql)
	if strings.HasPrefix(sql, "create") || strings.HasPrefix(sql, "CREATE") {
		return DDL
	} else if strings.HasPrefix(sql, "insert") || strings.HasPrefix(sql, "INSERT") ||
		strings.HasPrefix(sql, "delete") || strings.HasPrefix(sql, "DELETE") ||
		strings.HasPrefix(sql, "update") || strings.HasPrefix(sql, "UPDATE") ||
		strings.HasPrefix(sql, "select") || strings.HasPrefix(sql, "SELECT") {
		return DQL
	} else {
		return Unknown
	}
}

func ParseDDL(sql string) ([]Table, error) {
	sql = strings.TrimSpace(sql)
	if strings.HasPrefix(sql, "create") || strings.HasPrefix(sql, "CREATE") {
		return NewParser(NewLexer(sql).Tokenize()).Parse()
	} else {
		return nil, errors.New("传入ParseDDL中的不是create语句")
	}
}

func ParseDQL(sql []string) ([]Query, error) {
	return ParseMany(sql)
}
