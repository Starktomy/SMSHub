package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseSMSFrame_Valid(t *testing.T) {
	frame := `SMS_START:{"type":"sms","content":"hello"}:SMS_END`
	msg, err := parseSMSFrame(frame)
	if err != nil {
		t.Fatalf("解析有效帧失败: %v", err)
	}

	if msg.Type != "sms" {
		t.Errorf("类型解析错误，期望 sms，实际 %s", msg.Type)
	}
	if msg.Payload["content"] != "hello" {
		t.Errorf("Payload 解析错误")
	}
	if msg.JSON != `{"type":"sms","content":"hello"}` {
		t.Errorf("原始 JSON 提取错误")
	}
}

func TestParseSMSFrame_InvalidPrefix(t *testing.T) {
	frame := `INVALID:{"type":"sms"}:SMS_END`
	_, err := parseSMSFrame(frame)
	if err != errNotSMSFrame {
		t.Errorf("无效前缀应返回 errNotSMSFrame，实际返回 %v", err)
	}
}

func TestParseSMSFrame_InvalidSuffix(t *testing.T) {
	frame := `SMS_START:{"type":"sms"}:INVALID`
	_, err := parseSMSFrame(frame)
	if err != errNotSMSFrame {
		t.Errorf("无效后缀应返回 errNotSMSFrame，实际返回 %v", err)
	}
}

func TestParseSMSFrame_InvalidJSON(t *testing.T) {
	frame := `SMS_START:{invalid-json}:SMS_END`
	_, err := parseSMSFrame(frame)
	if err == nil {
		t.Error("无效 JSON 应返回错误")
	}
	if !strings.Contains(err.Error(), "JSON解析失败") {
		t.Errorf("错误信息应包含 JSON解析失败，实际: %v", err)
	}
}

func TestParseSMSFrame_MissingType(t *testing.T) {
	frame := `SMS_START:{"content":"hello"}:SMS_END`
	_, err := parseSMSFrame(frame)
	if err != errMissingType {
		t.Errorf("缺失 type 字段应返回 errMissingType，实际返回 %v", err)
	}
}

func TestBuildCommandMessage(t *testing.T) {
	cmd := map[string]string{
		"command": "send_sms",
		"to":      "10086",
	}

	bytes, jsonStr, err := buildCommandMessage(cmd)
	if err != nil {
		t.Fatalf("构建命令失败: %v", err)
	}

	// JSON 字段顺序可能不同，反序列化验证
	var decoded map[string]string
	json.Unmarshal([]byte(jsonStr), &decoded)
	if decoded["command"] != "send_sms" || decoded["to"] != "10086" {
		t.Errorf("JSON 内容不匹配")
	}

	expectedPrefix := "CMD_START:"
	expectedSuffix := ":CMD_END\r\n"
	result := string(bytes)

	if !strings.HasPrefix(result, expectedPrefix) {
		t.Errorf("结果应以 %s 开头", expectedPrefix)
	}
	if !strings.HasSuffix(result, expectedSuffix) {
		t.Errorf("结果应以 %s 结尾", expectedSuffix)
	}
}

func TestIsValidResponse(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{`{"type":"response"}`, true},
		{`{"timestamp":123456}`, true},
		{`{"key":"value"}`, true},
		{`invalid json`, false},
		{`SMS_START:raw-data:SMS_END`, true},
		{`system_ready`, true},
		{`heartbeat received`, true},
		{`random noise`, false},
	}

	for _, tt := range tests {
		if got := isValidResponse(tt.input); got != tt.want {
			t.Errorf("isValidResponse(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
