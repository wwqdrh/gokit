package dbparser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// {name: 1, age: 18} => "name = ?, age = ?", 1, 18
func GenQueryFromJson(data []byte) (string, []interface{}) {
	args := map[string]interface{}{}
	if err := json.Unmarshal(data, &args); err != nil {
		return "", nil
	}
	res := []interface{}{}
	parts := []string{}
	for key, val := range args {
		parts = append(parts, key+" = ?")
		res = append(res, val)
	}
	return strings.Join(parts, ","), res
}

func GenQueryFromQuery(data url.Values) (string, []interface{}) {
	res := []interface{}{}
	parts := []string{}
	for key, val := range data {
		parts = append(parts, key+" = ?")
		res = append(res, val[0])
	}
	return strings.Join(parts, ","), res
}

// CreateTableSQL 根据结构体和类型返回数据表创建语句
// for gormv2 model
func GenCreateTableStatement(model interface{}, tableName string, dbType string) (string, error) {
	// 获取结构体的类型和值
	t := reflect.TypeOf(model)
	// 检查是否是结构体
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("model is not a struct")
	}

	// 定义一个切片来存储列定义
	var columns []string
	// 遍历结构体的字段
	for i := 0; i < t.NumField(); i++ {
		// 获取字段的类型和标签
		field := t.Field(i)
		fieldType := field.Type
		fieldTag := field.Tag.Get("gorm")
		// 解析标签中的选项
		options := parseTagOptions(fieldTag)
		if _, ok := options["miss"]; ok {
			continue
		}

		// 如果有column选项，使用它作为列名，否则使用字段名
		columnName := options["column"]
		if columnName == "" {
			columnName = field.Name
		}
		// 如果有type选项，使用它作为列类型，否则根据字段类型和数据库类型推断列类型
		columnType := options["type"]
		if columnType == "" {
			columnType = inferColumnType(fieldType, dbType)
			if columnType == "" {
				return "", fmt.Errorf("cannot infer column type for %s", columnName)
			}
		}
		// 拼接列定义
		columnDef := fmt.Sprintf("%s %s", columnName, columnType)
		// 如果有primary_key选项，添加主键约束
		if _, ok := options["primaryKey"]; ok {
			columnDef += " PRIMARY KEY"
		}
		// 如果有not null选项，添加非空约束
		if _, ok := options["not null"]; ok {
			columnDef += " NOT NULL"
		}
		// 如果有unique_index选项，添加唯一索引约束
		if indexName, ok := options["unique_index"]; ok {
			columnDef += fmt.Sprintf(" UNIQUE INDEX %s", indexName)
		}
		// 如果有index选项，添加普通索引约束
		if indexName, ok := options["index"]; ok {
			columnDef += fmt.Sprintf(" INDEX %s", indexName)
		}
		// 如果有default选项，添加默认值约束
		if defaultValue, ok := options["default"]; ok {
			columnDef += fmt.Sprintf(" DEFAULT %s", defaultValue)
		}
		// 将列定义添加到切片中
		columns = append(columns, columnDef)
	}
	// 使用strings.Join将列定义用逗号分隔拼接成字符串
	columnsStr := strings.Join(columns, ", ")
	// 拼接SQL语句
	sql := fmt.Sprintf("CREATE TABLE %s (%s);", tableName, columnsStr)
	return sql, nil
}

// parseTagOptions 解析标签中的选项，返回一个映射表
func parseTagOptions(tag string) map[string]string {
	// 定义一个映射表来存储选项和值
	options := make(map[string]string)
	// 使用分号分隔标签中的选项
	parts := strings.Split(tag, ";")
	// 遍历每个选项
	for _, part := range parts {
		if part == "-" {
			options["miss"] = "true"
			return options
		}
		// 使用冒号分隔选项和值
		kv := strings.Split(part, ":")
		// 如果只有一个元素，说明没有值，使用空字符串作为值
		if len(kv) == 1 {
			options[kv[0]] = ""
		} else if len(kv) == 2 { // 如果有两个元素，说明有值，使用第二个元素作为值
			options[kv[0]] = kv[1]
		}
	}
	return options
}

// inferColumnType 根据字段类型和数据库类型推断列类型，返回一个字符串
func inferColumnType(fieldType reflect.Type, dbType string) string {
	// 定义一个映射表来存储不同的数据库类型对应的列类型
	var typeMap map[string]string
	switch dbType {
	case "sqlite":
		typeMap = map[string]string{
			"uint64":    "INTEGER",
			"string":    "TEXT",
			"uint":      "INTEGER",
			"int":       "INTEGER",
			"bigint":    "INTEGER",
			"time.Time": "DATETIME",
			"Time":      "DATETIME",
		}
	case "mysql":
		typeMap = map[string]string{
			"uint64":    "BIGINT UNSIGNED",
			"string":    "VARCHAR(255)",
			"uint":      "INT",
			"int":       "INT",
			"bigint":    "INT",
			"time.Time": "DATETIME",
			"Time":      "DATETIME",
		}
	default:
		return ""
	}
	// 根据字段类型的名称获取列类型，如果没有找到，返回空字符串
	columnType, ok := typeMap[fieldType.Name()]
	if !ok {
		return fieldType.Name()
	}
	return columnType
}

func GenListStatement(tableName string) string {
	return fmt.Sprintf("select * from %s limit :limit offset :offset", tableName)
}

func GenAddStatement(tableName string, columns []string) string {
	values := make([]string, len(columns))
	for i := range values {
		values[i] = ":" + columns[i]
	}
	return fmt.Sprintf("insert into %s (%s) values (%s)", tableName, strings.Join(columns, ","), strings.Join(values, ","))
}

func GenGetStatement(tableName, pkName string) string {
	return fmt.Sprintf("select * from %s where %s = :%s", tableName, pkName, pkName)
}

func GenDeleteStatement(tableName, pkName string) string {
	return fmt.Sprintf("delete from %s where %s = :%s", tableName, pkName, pkName)
}

func GenUpdateStatement(tableName, pkName string, updateColumns []string) string {
	updates := []string{}
	for _, item := range updateColumns {
		updates = append(updates, fmt.Sprintf("%s = :%s", item, item))
	}
	return fmt.Sprintf("update %s set %s where %s = :%s", tableName, strings.Join(updates, ","), pkName, pkName)
}
