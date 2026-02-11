package models

// Device status constants
const (
	DeviceStatusOnline  = "online"  // 心跳正常
	DeviceStatusOffline = "offline" // 心跳超时 (>60s)
	DeviceStatusError   = "error"   // 串口错误
)

// Device 设备模型
type Device struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`                     // 设备名称，如 "香港卡1"
	SerialPort  string `gorm:"unique" json:"serialPort"` // 串口路径 /dev/ttyUSB0
	Status      string `json:"status"`                   // online/offline/error
	PhoneNumber string `json:"phoneNumber"`              // SIM卡号码
	IMSI        string `json:"imsi"`                     // IMSI
	ICCID       string `json:"iccid"`                    // ICCID
	Operator    string `json:"operator"`                 // 运营商
	SignalLevel int    `json:"signalLevel"`              // 信号强度
	Flymode     bool   `json:"flymode"`                  // 飞行模式
	Enabled     bool   `json:"enabled"`                  // 是否启用
	GroupName   string `json:"groupName"`                // 设备分组
	LastSeenAt  int64  `json:"lastSeenAt"`               // 最后心跳时间
	CreatedAt   int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt   int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}
