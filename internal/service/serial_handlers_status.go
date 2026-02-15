package service

import (
	"encoding/json"

	"go.uber.org/zap"
)

type StatusData struct {
	Flymode bool   `json:"flymode"` // 兼容性保留（根级别）
	Type    string `json:"type"`    // 消息类型
	Version string `json:"version"` // Lua 脚本版本
	Iccid   string `json:"iccid"`   // 兼容性保留（根级别）
	Mobile  struct {
		IsRegistered bool    `json:"is_registered"`
		IsRoaming    bool    `json:"is_roaming"`
		Iccid        string  `json:"iccid"`
		SignalDesc   string  `json:"signal_desc"`
		SignalLevel  int     `json:"signal_level"`
		SimReady     bool    `json:"sim_ready"`
		Rssi         int     `json:"rssi"`
		Csq          int     `json:"csq"`          // CSQ 信号强度 (0-31)
		Rsrp         int     `json:"rsrp"`         // 参考信号接收功率 (-44 到 -140)
		Rsrq         float64 `json:"rsrq"`         // 参考信号发送功率 (-3 到 -19.5)
		Imsi         string  `json:"imsi"`         // SIM 卡 IMSI
		Number       string  `json:"number"`       // 手机号
		Operator     string  `json:"operator"`     // 运营商名称 (当前网络)
		SimOperator  string  `json:"sim_operator"` // 运营商名称 (SIM卡归属)
		Mnc          string  `json:"mnc"`          // 实时网络 MNC/PLMN (来自 Lua)
		Lac          int     `json:"lac"`          // 实时网络 LAC/TAC
		Cid          int     `json:"cid"`          // 实时网络 CellID
		Uptime       int64   `json:"uptime"`       // 模块开机时长，单位为秒
		Flymode      bool    `json:"flymode"`      // 飞行模式状态
	} `json:"mobile"`
	Timestamp int    `json:"timestamp"`
	MemKb     int    `json:"mem_kb"`
	PortName  string `json:"port_name"` // 串口名称
	Connected bool   `json:"connected"` // 连接状态
}

func (s *SerialService) handleStatusResponse(msg *ParsedMessage) {
	var statusData StatusData
	if err := json.Unmarshal([]byte(msg.JSON), &statusData); err != nil {
		s.logger.Error("JSON解析失败", zap.Error(err), zap.String("data", msg.JSON))
		return
	}

	// 1. 处理飞行模式（逻辑取反：硬件返回 true 代表关闭）
	realFlymode := !statusData.Mobile.Flymode
	s.flyMode.Store(realFlymode)
	statusData.Flymode = realFlymode

	// 2. 处理 ICCID（直接从 mobile 中提取，不进行复杂保护）
	statusData.Iccid = statusData.Mobile.Iccid

	// 3. 处理运营商
	// (A) SIM卡归属运营商 - 基于 IMSI
	if len(statusData.Mobile.Imsi) >= 5 {
		simPlmn := statusData.Mobile.Imsi[:5]
		statusData.Mobile.SimOperator = func() string {
			if v, ok := OperData[simPlmn]; ok {
				return v
			}
			return simPlmn
		}()
	}

	// (B) 实时网络运营商 - 基于基站 MNC (Lua提供)
	if statusData.Mobile.Mnc != "" {
		netPlmn := statusData.Mobile.Mnc
		statusData.Mobile.Operator = func() string {
			if v, ok := OperData[netPlmn]; ok {
				return v
			}
			// 尝试前 5 位匹配 (处理 3 位 MNC 的情况)
			if len(netPlmn) > 5 {
				if v, ok := OperData[netPlmn[:5]]; ok {
					return v
				}
			}
			return netPlmn
		}()
	} else {
		// 如果没有实时基站信息，则 Operator 显示为与 SIM 卡一致
		statusData.Mobile.Operator = statusData.Mobile.SimOperator
	}

	s.deviceCache.Set(CacheKeyDeviceStatus, &statusData, CacheTTL)
	if s.statusUpdateCallback != nil {
		s.statusUpdateCallback(&statusData)
	}
	s.logger.Debug("设备状态缓存已更新")
}

func (s *SerialService) handleSystemReady(msg *ParsedMessage) {
	if message, ok := msg.Payload["message"].(string); ok {
		s.logger.Info("系统就绪", zap.String("message", message))
	}
}

func (s *SerialService) handleHeartbeat(msg *ParsedMessage) {
	timestamp, _ := msg.Payload["timestamp"].(float64)
	memoryUsage, _ := msg.Payload["memory_usage"].(float64)
	bufferSize, _ := msg.Payload["buffer_size"].(float64)

	// 触发状态更新回调，仅更新在线时间，不传递业务数据
	if s.statusUpdateCallback != nil {
		s.statusUpdateCallback(nil)
	}

	s.logger.Debug("设备心跳",
		zap.Int64("timestamp", int64(timestamp)),
		zap.Float64("memory_usage", memoryUsage),
		zap.Int("buffer_size", int(bufferSize)))
}

func (s *SerialService) handleCellularControlResponse(msg *ParsedMessage) {
	s.logger.Debug("收到蜂窝网络控制响应", zap.Any("data", msg.Payload))
}

func (s *SerialService) handlePhoneNumberResponse(msg *ParsedMessage) {
	s.logger.Debug("收到电话号码响应", zap.Any("data", msg.Payload))
}

func (s *SerialService) handleCommandResponse(msg *ParsedMessage) {
	if action, ok := msg.Payload["action"].(string); ok {
		s.logger.Info("命令响应", zap.String("action", action), zap.Any("result", msg.Payload["result"]))
	}
}

func (s *SerialService) handleSIMEvent(msg *ParsedMessage) {
	status, _ := msg.Payload["status"].(string)
	s.logger.Info("SIM卡事件", zap.String("status", status))
}

func (s *SerialService) handleWarningMessage(msg *ParsedMessage) {
	if warnMsg, ok := msg.Payload["msg"].(string); ok {
		s.logger.Warn("设备警告", zap.String("message", warnMsg))
	}
}

func (s *SerialService) handleErrorMessage(msg *ParsedMessage) {
	if errMsg, ok := msg.Payload["msg"].(string); ok {
		s.logger.Error("设备错误", zap.String("message", errMsg))
	}
}
