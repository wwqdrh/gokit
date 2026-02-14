package dbparser

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

func TestGetTablesFromSQL(t *testing.T) {
	tables, err := GetTables(`
-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uname TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    nickname TEXT NOT NULL,
    roleid INTEGER NOT NULL DEFAULT 2,
    email TEXT NOT NULL UNIQUE,
    phone TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建设备分组表
CREATE TABLE IF NOT EXISTS device_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    device_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建设备表
CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    brand TEXT NOT NULL,
    model TEXT NOT NULL,
    ip TEXT NOT NULL,
    port INTEGER NOT NULL,
    status TEXT NOT NULL,
    online_time TIMESTAMP,
    group_id TEXT,
    location TEXT NOT NULL,
    description TEXT,
    last_active TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建设备扩展信息表
CREATE TABLE IF NOT EXISTS device_extra (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES devices(id),
    UNIQUE(device_id, key)
);

-- 创建告警表
CREATE TABLE IF NOT EXISTS alarms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    device_name TEXT NOT NULL,
    device_type TEXT NOT NULL,
    level TEXT NOT NULL,
    status TEXT NOT NULL,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    trigger_time TIMESTAMP NOT NULL,
    process_time TIMESTAMP,
    handler TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES devices(id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_devices_group_id ON devices(group_id);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type);
CREATE INDEX IF NOT EXISTS idx_alarms_device_id ON alarms(device_id);
CREATE INDEX IF NOT EXISTS idx_alarms_level ON alarms(level);
CREATE INDEX IF NOT EXISTS idx_alarms_status ON alarms(status);
CREATE INDEX IF NOT EXISTS idx_alarms_trigger_time ON alarms(trigger_time);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- 插入默认角色数据
INSERT OR IGNORE INTO roles (name, code, description, permissions) VALUES
('管理员', 'admin', '拥有所有权限', 'user:manage,device:manage,alarm:manage,statistics:view,settings:manage'),
('操作员', 'operator', '可以管理设备和处理告警', 'device:manage,alarm:manage,statistics:view'),
('查看员', 'viewer', '只能查看数据', 'statistics:view');

-- 插入默认用户数据
INSERT OR IGNORE INTO users (uname, password, nickname, roleid, email, phone, status) VALUES
('admin', '[[password]]123456', '管理员', 1, 'admin@example.com', '13800138000', 'active'),
('user1', '[[password]]123456', '用户1', 2, 'user1@example.com', '13800138001', 'active'),
('user2', '[[password]]123456', '用户2', 3, 'user2@example.com', '13800138002', 'inactive');

-- 插入默认设备分组数据
INSERT OR IGNORE INTO device_groups (name, description, device_count) VALUES
('摄像头组', '所有摄像头设备', 4),
('录像机组', '所有录像机设备', 2),
('门禁组', '所有门禁系统设备', 2);

-- 插入默认设备数据
INSERT OR IGNORE INTO devices (name, type, brand, model, ip, port, status, online_time, group_id, location, description, last_active) VALUES
('海康摄像头-001', 'camera', 'hikvision', 'DS-2CD2143G0-I', '192.168.1.101', 80, 'online', '2024-01-01T00:00:00Z', 'group-001', '一号楼一层大厅', '海康威视400万像素网络摄像头', '2024-01-01T12:00:00Z'),
('大华摄像头-002', 'camera', 'dahua', 'DH-IPC-HFW4431M-I1', '192.168.1.102', 80, 'online', '2024-01-01T00:00:00Z', 'group-001', '一号楼二层走廊', '大华400万像素网络摄像头', '2024-01-01T12:00:00Z'),
('海康录像机-001', 'dvr', 'hikvision', 'DS-7816N-K2', '192.168.1.201', 80, 'online', '2024-01-01T00:00:00Z', 'group-002', '监控室', '海康威视16路网络硬盘录像机', '2024-01-01T12:00:00Z'),
('门禁系统-001', 'access_control', 'hikvision', 'DS-K1T601M', '192.168.1.301', 80, 'online', '2024-01-01T00:00:00Z', 'group-003', '一号楼入口', '海康威视人脸识别门禁一体机', '2024-01-01T12:00:00Z'),
('摄像头-003', 'camera', 'other', 'Generic-Camera-001', '192.168.1.103', 80, 'offline', NULL, 'group-001', '二号楼一层大厅', '通用网络摄像头', '2024-01-01T10:00:00Z'),
('门禁系统-002', 'access_control', 'dahua', 'DH-ASG2113B', '192.168.1.302', 80, 'warning', '2024-01-01T00:00:00Z', 'group-003', '二号楼入口', '大华刷卡门禁系统', '2024-01-01T11:00:00Z'),
('海康摄像头-004', 'camera', 'hikvision', 'DS-2CD2T43G0-I5', '192.168.1.104', 80, 'online', '2024-01-01T00:00:00Z', 'group-001', '三号楼一层大厅', '海康威视400万像素红外摄像头', '2024-01-01T12:00:00Z'),
('大华录像机-002', 'dvr', 'dahua', 'DH-NVR4416-HDS2', '192.168.1.202', 80, 'error', NULL, 'group-002', '备用监控室', '大华16路网络硬盘录像机', '2024-01-01T09:00:00Z');

-- 插入默认设备扩展信息数据
INSERT OR IGNORE INTO device_extra (device_id, key, value) VALUES
('device-003', 'storageCapacity', '4TB'),
('device-003', 'usedStorage', '2.5TB'),
('device-004', 'doorStatus', 'closed'),
('device-006', 'doorStatus', 'closed'),
('device-008', 'storageCapacity', '8TB'),
('device-008', 'usedStorage', '6TB');

-- 插入默认告警数据
INSERT OR IGNORE INTO alarms (device_id, device_name, device_type, level, status, type, content, trigger_time, process_time, handler) VALUES
('device-005', '摄像头-003', '摄像头', 'error', 'unprocessed', '设备离线', '设备长时间未响应，可能已离线', '2024-01-01 10:00:00', NULL, NULL),
('device-006', '门禁系统-002', '门禁系统', 'warning', 'processing', '门禁异常', '门禁刷卡失败次数过多', '2024-01-01 11:00:00', NULL, NULL),
('device-008', '大华录像机-002', '录像机', 'critical', 'unprocessed', '设备故障', '设备无法连接，可能硬件故障', '2024-01-01 09:00:00', NULL, NULL),
('device-003', '海康录像机-001', '录像机', 'info', 'processed', '存储预警', '存储空间使用超过80%', '2024-01-01 08:00:00', '2024-01-01 08:30:00', '管理员'),
('device-001', '海康摄像头-001', '摄像头', 'warning', 'processed', '网络异常', '网络延迟过高', '2024-01-01 07:00:00', '2024-01-01 07:15:00', '管理员'),
('device-004', '门禁系统-001', '门禁系统', 'info', 'processed', '门禁开关', '门禁正常开启和关闭', '2024-01-01 06:00:00', '2024-01-01 06:00:00', '系统'),
('device-002', '大华摄像头-002', '摄像头', 'warning', 'unprocessed', '画面异常', '画面模糊，可能镜头被遮挡', '2024-01-01 12:00:00', NULL, NULL),
('device-007', '海康摄像头-004', '摄像头', 'info', 'processed', '设备上线', '设备正常上线', '2024-01-01 00:00:00', '2024-01-01 00:00:00', '系统');
	`)
	if err != nil {
		t.Error(err)
	}
	if len(tables) != 6 {
		t.Error(errors.New("数据表个数不对"))
	}
	if tables[0].Name != "roles" {
		t.Error()
	}
	if len(tables[0].Columns) != 7 {
		t.Error()
	}
	if tables[0].Columns[1].Type != "TEXT" {
		t.Error()
	}
}
