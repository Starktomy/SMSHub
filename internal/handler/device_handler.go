package handler

import (
	"net/http"

	"github.com/Starktomy/smshub/internal/models"
	"github.com/Starktomy/smshub/internal/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// DeviceHandler 设备管理API处理器
type DeviceHandler struct {
	logger        *zap.Logger
	deviceManager *service.DeviceManager
}

// NewDeviceHandler 创建设备Handler实例
func NewDeviceHandler(logger *zap.Logger, deviceManager *service.DeviceManager) *DeviceHandler {
	return &DeviceHandler{
		logger:        logger,
		deviceManager: deviceManager,
	}
}

// List 获取设备列表
// GET /api/devices
func (h *DeviceHandler) List(c echo.Context) error {
	devices, err := h.deviceManager.GetAllDevices(c.Request().Context())
	if err != nil {
		h.logger.Error("获取设备列表失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取设备列表失败",
		})
	}
	return c.JSON(http.StatusOK, devices)
}

// Get 获取单个设备
// GET /api/devices/:id
func (h *DeviceHandler) Get(c echo.Context) error {
	id := c.Param("id")
	device, err := h.deviceManager.GetDevice(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "设备不存在",
		})
	}
	return c.JSON(http.StatusOK, device)
}

// CreateDeviceRequest 创建设备请求
type CreateDeviceRequest struct {
	Name       string `json:"name"`
	SerialPort string `json:"serialPort"`
	GroupName  string `json:"groupName"`
	Enabled    bool   `json:"enabled"`
}

// Create 添加设备
// POST /api/devices
func (h *DeviceHandler) Create(c echo.Context) error {
	var req CreateDeviceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "请求参数错误",
		})
	}

	if req.SerialPort == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "串口路径不能为空",
		})
	}

	device := &models.Device{
		Name:       req.Name,
		SerialPort: req.SerialPort,
		GroupName:  req.GroupName,
		Enabled:    req.Enabled,
	}

	if err := h.deviceManager.CreateDevice(c.Request().Context(), device); err != nil {
		h.logger.Error("创建设备失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "创建设备失败",
		})
	}

	return c.JSON(http.StatusOK, device)
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	Name       string `json:"name"`
	SerialPort string `json:"serialPort"`
	GroupName  string `json:"groupName"`
	Enabled    bool   `json:"enabled"`
}

// Update 更新设备信息
// PUT /api/devices/:id
func (h *DeviceHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateDeviceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "请求参数错误",
		})
	}

	device := &models.Device{
		ID:         id,
		Name:       req.Name,
		SerialPort: req.SerialPort,
		GroupName:  req.GroupName,
		Enabled:    req.Enabled,
	}

	if err := h.deviceManager.UpdateDevice(c.Request().Context(), device); err != nil {
		h.logger.Error("更新设备失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "更新设备失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "更新成功",
	})
}

// Delete 删除设备
// DELETE /api/devices/:id
func (h *DeviceHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.deviceManager.DeleteDevice(c.Request().Context(), id); err != nil {
		h.logger.Error("删除设备失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "删除设备失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "删除成功",
	})
}

// Enable 启用设备
// POST /api/devices/:id/enable
func (h *DeviceHandler) Enable(c echo.Context) error {
	id := c.Param("id")
	if err := h.deviceManager.EnableDevice(c.Request().Context(), id); err != nil {
		h.logger.Error("启用设备失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "启用设备失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "启用成功",
	})
}

// Disable 禁用设备
// POST /api/devices/:id/disable
func (h *DeviceHandler) Disable(c echo.Context) error {
	id := c.Param("id")
	if err := h.deviceManager.DisableDevice(c.Request().Context(), id); err != nil {
		h.logger.Error("禁用设备失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "禁用设备失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "禁用成功",
	})
}

// SetFlymodeRequest 设置飞行模式请求
type SetDeviceFlymodeRequest struct {
	Enabled bool `json:"enabled"`
}

// SetFlymode 设置飞行模式
// POST /api/devices/:id/flymode
func (h *DeviceHandler) SetFlymode(c echo.Context) error {
	id := c.Param("id")
	var req SetDeviceFlymodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "请求参数错误",
		})
	}

	if err := h.deviceManager.SetDeviceFlymode(c.Request().Context(), id, req.Enabled); err != nil {
		h.logger.Error("设置飞行模式失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "设置飞行模式失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "设置成功",
	})
}

// Reboot 重启设备
// POST /api/devices/:id/reboot
func (h *DeviceHandler) Reboot(c echo.Context) error {
	id := c.Param("id")
	if err := h.deviceManager.RebootDevice(c.Request().Context(), id); err != nil {
		h.logger.Error("重启设备失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "重启设备失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "重启成功",
	})
}

// GetStatus 获取设备状态
// GET /api/devices/:id/status
func (h *DeviceHandler) GetStatus(c echo.Context) error {
	id := c.Param("id")
	status, err := h.deviceManager.GetDeviceStatus(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取设备状态失败",
		})
	}

	return c.JSON(http.StatusOK, status)
}

// Discover 扫描可用串口
// GET /api/devices/discover
func (h *DeviceHandler) Discover(c echo.Context) error {
	ports, err := h.deviceManager.DiscoverSerialPorts()
	if err != nil {
		h.logger.Error("扫描串口失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "扫描串口失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ports": ports,
	})
}

// GetGroups 获取设备分组列表
// GET /api/devices/groups
func (h *DeviceHandler) GetGroups(c echo.Context) error {
	groups, err := h.deviceManager.GetDeviceGroups(c.Request().Context())
	if err != nil {
		h.logger.Error("获取设备分组失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取设备分组失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"groups": groups,
	})
}

// SendDeviceSMSRequest 指定设备发送短信请求
type SendDeviceSMSRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

// SendSMS 指定设备发送短信
// POST /api/devices/:id/sms
func (h *DeviceHandler) SendSMS(c echo.Context) error {
	id := c.Param("id")
	var req SendDeviceSMSRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "请求参数错误",
		})
	}

	if req.To == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "手机号和内容不能为空",
		})
	}

	msgID, err := h.deviceManager.SendSMSByDevice(id, req.To, req.Content)
	if err != nil {
		h.logger.Error("发送短信失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "发送短信失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":   "发送成功",
		"messageId": msgID,
	})
}

// AutoSendSMSRequest 自动选择设备发送短信请求
type AutoSendSMSRequest struct {
	To       string               `json:"to"`
	Content  string               `json:"content"`
	Strategy service.SendStrategy `json:"strategy"`
}

// AutoSendSMS 自动选择设备发送短信
// POST /api/sms/send
func (h *DeviceHandler) AutoSendSMS(c echo.Context) error {
	var req AutoSendSMSRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "请求参数错误",
		})
	}

	if req.To == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "手机号和内容不能为空",
		})
	}

	if req.Strategy == "" {
		req.Strategy = service.StrategyAuto
	}

	msgID, deviceID, err := h.deviceManager.SendSMS(req.To, req.Content, req.Strategy)
	if err != nil {
		h.logger.Error("发送短信失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "发送短信失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":   "发送成功",
		"messageId": msgID,
		"deviceId":  deviceID,
	})
}

// BatchSendSMS 批量发送短信
// POST /api/sms/batch
func (h *DeviceHandler) BatchSendSMS(c echo.Context) error {
	var req service.BatchSendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "请求参数错误",
		})
	}

	if len(req.Recipients) == 0 || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "收件人列表和内容不能为空",
		})
	}

	if req.Strategy == "" {
		req.Strategy = service.StrategyAuto
	}

	results := h.deviceManager.BatchSendSMS(&req)

	return c.JSON(http.StatusOK, map[string]any{
		"results": results,
	})
}

// GetStats 获取设备统计信息
// GET /api/devices/stats
func (h *DeviceHandler) GetStats(c echo.Context) error {
	stats, err := h.deviceManager.GetDeviceStats(c.Request().Context())
	if err != nil {
		h.logger.Error("获取设备统计失败", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取设备统计失败",
		})
	}

	return c.JSON(http.StatusOK, stats)
}
