package onvif

// import (
// 	"context"
// 	"fmt"

// 	"github.com/use-go/onvif"
// 	"github.com/use-go/onvif/device"
// 	"github.com/use-go/onvif/media"
// 	"github.com/use-go/onvif/ptz"
// )

// // Client 是 ONVIF 客户端的封装
// type Client struct {
// 	device *onvif.Device
// 	ctx    context.Context
// }

// // NewClient 创建一个新的 ONVIF 客户端
// func NewClient(address, username, password string) (*Client, error) {
// 	ctx := context.Background()
// 	device, err := onvif.NewDevice(onvif.DeviceParams{
// 		Xaddr:  address,
// 		Username: username,
// 		Password: password,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create device: %v", err)
// 	}

// 	return &Client{
// 		device: device,
// 		ctx:    ctx,
// 	}, nil
// }

// // GetDeviceInformation 获取设备信息
// func (c *Client) GetDeviceInformation() (*device.GetDeviceInformationResponse, error) {
// 	resp, err := c.device.CallMethod(c.ctx, device.GetDeviceInformation{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get device information: %v", err)
// 	}

// 	deviceInfo, ok := resp.(*device.GetDeviceInformationResponse)
// 	if !ok {
// 		return nil, fmt.Errorf("failed to cast response to GetDeviceInformationResponse")
// 	}

// 	return deviceInfo, nil
// }

// // GetProfiles 获取设备的媒体配置文件
// func (c *Client) GetProfiles() (*media.GetProfilesResponse, error) {
// 	resp, err := c.device.CallMethod(c.ctx, media.GetProfiles{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get profiles: %v", err)
// 	}

// 	profiles, ok := resp.(*media.GetProfilesResponse)
// 	if !ok {
// 		return nil, fmt.Errorf("failed to cast response to GetProfilesResponse")
// 	}

// 	return profiles, nil
// }

// // GetStreamURI 获取媒体流的 URI
// func (c *Client) GetStreamURI(profileToken string) (*media.GetStreamUriResponse, error) {
// 	resp, err := c.device.CallMethod(c.ctx, media.GetStreamUri{
// 		ProfileToken: profileToken,
// 		StreamSetup: media.StreamSetup{
// 			Stream:    "RTP-Unicast",
// 			Transport: media.Transport{Protocol: "RTSP"},
// 		},
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get stream URI: %v", err)
// 	}

// 	streamURI, ok := resp.(*media.GetStreamUriResponse)
// 	if !ok {
// 		return nil, fmt.Errorf("failed to cast response to GetStreamUriResponse")
// 	}

// 	return streamURI, nil
// }

// // PTZAbsoluteMove 执行 PTZ 绝对移动
// func (c *Client) PTZAbsoluteMove(profileToken string, position ptz.Position, speed ptz.Velocity) error {
// 	_, err := c.device.CallMethod(c.ctx, ptz.AbsoluteMove{
// 		ProfileToken: profileToken,
// 		Position:     position,
// 		Speed:        speed,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to perform PTZ absolute move: %v", err)
// 	}

// 	return nil
// }

// // PTZRelativeMove 执行 PTZ 相对移动
// func (c *Client) PTZRelativeMove(profileToken string, translation ptz.Translation, speed ptz.Velocity) error {
// 	_, err := c.device.CallMethod(c.ctx, ptz.RelativeMove{
// 		ProfileToken: profileToken,
// 		Translation:  translation,
// 		Speed:        speed,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to perform PTZ relative move: %v", err)
// 	}

// 	return nil
// }

// // PTZStop 停止 PTZ 移动
// func (c *Client) PTZStop(profileToken string) error {
// 	_, err := c.device.CallMethod(c.ctx, ptz.Stop{
// 		ProfileToken: profileToken,
// 		PanTilt:      true,
// 		Zoom:         true,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to stop PTZ movement: %v", err)
// 	}

// 	return nil
// }

// // StreamType 定义流类型
// type StreamType string

// const (
// 	// StreamTypeRTSP RTSP 流
// 	StreamTypeRTSP StreamType = "RTSP"
// 	// StreamTypeHTTP HTTP 流
// 	StreamTypeHTTP StreamType = "HTTP"
// 	// StreamTypeHLS HLS 流 (HTTP Live Streaming)
// 	StreamTypeHLS StreamType = "HLS"
// 	// StreamTypeWebRTC WebRTC 流
// 	StreamTypeWebRTC StreamType = "WebRTC"
// )

// // StreamInfo 流信息
// type StreamInfo struct {
// 	// 流地址
// 	URI string
// 	// 流类型
// 	Type StreamType
// 	// 媒体配置文件令牌
// 	ProfileToken string
// }

// // GetCameraStream 获取摄像头流信息
// func (c *Client) GetCameraStream(profileToken string) (*StreamInfo, error) {
// 	// 获取 RTSP 流地址
// 	rtspResp, err := c.GetStreamURI(profileToken)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get RTSP stream: %v", err)
// 	}

// 	// 构造流信息
// 	streamInfo := &StreamInfo{
// 		URI:          rtspResp.MediaUri.Uri,
// 		Type:         StreamTypeRTSP,
// 		ProfileToken: profileToken,
// 	}

// 	return streamInfo, nil
// }

// // GetDefaultCameraStream 获取默认摄像头流信息
// func (c *Client) GetDefaultCameraStream() (*StreamInfo, error) {
// 	// 获取所有媒体配置文件
// 	profiles, err := c.GetProfiles()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get profiles: %v", err)
// 	}

// 	// 检查是否有配置文件
// 	if len(profiles.Profiles) == 0 {
// 		return nil, fmt.Errorf("no profiles found")
// 	}

// 	// 使用第一个配置文件
// 	defaultProfile := profiles.Profiles[0]

// 	// 获取流信息
// 	return c.GetCameraStream(defaultProfile.Token)
// }
