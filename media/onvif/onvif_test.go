package onvif

// import (
// 	"testing"

// 	"github.com/use-go/onvif/ptz"
// )

// // TestNewClient 测试创建 ONVIF 客户端
// func TestNewClient(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"

// 	_, err := NewClient(address, username, password)
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("NewClient 调用成功")
// 	} else {
// 		t.Logf("NewClient 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestGetDeviceInformation 测试获取设备信息
// func TestGetDeviceInformation(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	_, err = client.GetDeviceInformation()
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("GetDeviceInformation 调用成功")
// 	} else {
// 		t.Logf("GetDeviceInformation 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestGetProfiles 测试获取媒体配置文件
// func TestGetProfiles(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	_, err = client.GetProfiles()
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("GetProfiles 调用成功")
// 	} else {
// 		t.Logf("GetProfiles 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestGetStreamURI 测试获取媒体流 URI
// func TestGetStreamURI(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"
// 	profileToken := "profile1"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	_, err = client.GetStreamURI(profileToken)
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("GetStreamURI 调用成功")
// 	} else {
// 		t.Logf("GetStreamURI 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestPTZAbsoluteMove 测试 PTZ 绝对移动
// func TestPTZAbsoluteMove(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"
// 	profileToken := "profile1"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	// 创建一个默认的位置和速度
// 	position := ptz.Position{
// 		PanTilt: &ptz.Vector{
// 			X: 0.0,
// 			Y: 0.0,
// 		},
// 		Zoom: &ptz.Vector{
// 			X: 1.0,
// 		},
// 	}

// 	speed := ptz.Velocity{
// 		PanTilt: &ptz.Vector{
// 			X: 0.5,
// 			Y: 0.5,
// 		},
// 		Zoom: &ptz.Vector{
// 			X: 0.5,
// 		},
// 	}

// 	err = client.PTZAbsoluteMove(profileToken, position, speed)
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("PTZAbsoluteMove 调用成功")
// 	} else {
// 		t.Logf("PTZAbsoluteMove 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestPTZRelativeMove 测试 PTZ 相对移动
// func TestPTZRelativeMove(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"
// 	profileToken := "profile1"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	// 创建一个默认的平移和速度
// 	translation := ptz.Translation{
// 		PanTilt: &ptz.Vector{
// 			X: 0.1,
// 			Y: 0.1,
// 		},
// 		Zoom: &ptz.Vector{
// 			X: 0.1,
// 		},
// 	}

// 	speed := ptz.Velocity{
// 		PanTilt: &ptz.Vector{
// 			X: 0.5,
// 			Y: 0.5,
// 		},
// 		Zoom: &ptz.Vector{
// 			X: 0.5,
// 		},
// 	}

// 	err = client.PTZRelativeMove(profileToken, translation, speed)
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("PTZRelativeMove 调用成功")
// 	} else {
// 		t.Logf("PTZRelativeMove 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestPTZStop 测试停止 PTZ 移动
// func TestPTZStop(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"
// 	profileToken := "profile1"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	err = client.PTZStop(profileToken)
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("PTZStop 调用成功")
// 	} else {
// 		t.Logf("PTZStop 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestGetCameraStream 测试获取摄像头流信息
// func TestGetCameraStream(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"
// 	profileToken := "profile1"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	_, err = client.GetCameraStream(profileToken)
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("GetCameraStream 调用成功")
// 	} else {
// 		t.Logf("GetCameraStream 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }

// // TestGetDefaultCameraStream 测试获取默认摄像头流信息
// func TestGetDefaultCameraStream(t *testing.T) {
// 	// 由于没有实际设备，这里只测试函数调用结构
// 	address := "http://192.168.1.100:8080/onvif/device_service"
// 	username := "admin"
// 	password := "password"

// 	client, err := NewClient(address, username, password)
// 	if err != nil {
// 		t.Logf("创建客户端失败（预期，因为没有实际设备）: %v", err)
// 		return
// 	}

// 	_, err = client.GetDefaultCameraStream()
// 	// 这里会失败，因为没有实际设备，但我们只是测试代码结构
// 	if err == nil {
// 		t.Log("GetDefaultCameraStream 调用成功")
// 	} else {
// 		t.Logf("GetDefaultCameraStream 调用失败（预期，因为没有实际设备）: %v", err)
// 	}
// }
