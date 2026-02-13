package models

// Device status constants
const (
	DeviceStatusOnline  = "online"  // 心跳正常
	DeviceStatusOffline = "offline" // 心跳超时 (>60s)
	DeviceStatusError   = "error"   // 串口错误
)

// Device 设备模型
type Device struct {
	ID          string `gorm:"primaryKey;column:id" json:"id"`
	Name        string `gorm:"column:name" json:"name"`                     // 设备名称，如 "香港卡1"
	SerialPort  string `gorm:"unique;column:serial_port" json:"serialPort"` // 串口路径 /dev/ttyUSB0
	Status      string `gorm:"column:status" json:"status"`                 // online/offline/error
	PhoneNumber string `gorm:"column:phone_number" json:"phoneNumber"`      // SIM卡号码
	IMSI        string `gorm:"column:imsi" json:"imsi"`                     // IMSI
	ICCID       string `gorm:"column:iccid" json:"iccid"`                   // ICCID
	Operator    string `gorm:"column:operator" json:"operator"`             // 当前网络运营商
	SimOperator string `gorm:"column:sim_operator" json:"simOperator"`      // SIM卡所属运营商
	SignalLevel int    `gorm:"column:signal_level" json:"signalLevel"`      // 信号强度
	Flymode     bool   `gorm:"column:flymode" json:"flymode"`               // 飞行模式
	Enabled     bool   `gorm:"column:enabled" json:"enabled"`               // 是否启用
	GroupName   string `gorm:"column:group_name" json:"groupName"`          // 设备分组
	LastSeenAt  int64  `gorm:"column:last_seen_at" json:"lastSeenAt"`       // 最后心跳时间
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:milli" json:"createdAt"`
	UpdatedAt   int64  `gorm:"column:updated_at;autoUpdateTime:milli" json:"updatedAt"`
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}
