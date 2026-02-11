package service

import (
	"sync"
	"testing"
	"time"
)

func TestOIDCStateStoreConcurrency(t *testing.T) {
	svc := &OIDCService{
		stateStore: make(map[string]time.Time),
	}

	// 并发写入 state
	var wg sync.WaitGroup
	states := make([]string, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			svc.stateMu.Lock()
			state := "state-" + time.Now().String() + "-" + string(rune(idx))
			svc.stateStore[state] = time.Now().Add(10 * time.Minute)
			states[idx] = state
			svc.stateMu.Unlock()
		}(i)
	}
	wg.Wait()

	// 并发读取和删除
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if states[idx] != "" {
				svc.validateAndDeleteState(states[idx])
			}
		}(i)
	}
	wg.Wait()
}

func TestOIDCValidateAndDeleteState(t *testing.T) {
	svc := &OIDCService{
		stateStore: make(map[string]time.Time),
	}

	// 添加一个有效 state
	svc.stateStore["valid-state"] = time.Now().Add(10 * time.Minute)
	// 添加一个过期 state
	svc.stateStore["expired-state"] = time.Now().Add(-1 * time.Minute)

	// 验证有效 state
	if !svc.validateAndDeleteState("valid-state") {
		t.Error("应该验证通过有效的 state")
	}

	// 验证 state 已被删除
	if svc.validateAndDeleteState("valid-state") {
		t.Error("已使用的 state 不应再次验证通过")
	}

	// 验证过期 state
	if svc.validateAndDeleteState("expired-state") {
		t.Error("过期的 state 不应验证通过")
	}

	// 验证不存在的 state
	if svc.validateAndDeleteState("nonexistent") {
		t.Error("不存在的 state 不应验证通过")
	}
}

func TestOIDCCleanExpiredStates(t *testing.T) {
	svc := &OIDCService{
		stateStore: make(map[string]time.Time),
	}

	svc.stateStore["valid"] = time.Now().Add(10 * time.Minute)
	svc.stateStore["expired1"] = time.Now().Add(-1 * time.Minute)
	svc.stateStore["expired2"] = time.Now().Add(-5 * time.Minute)

	svc.cleanExpiredStates()

	if len(svc.stateStore) != 1 {
		t.Errorf("清理后应该只剩 1 个 state，实际有 %d 个", len(svc.stateStore))
	}

	if _, exists := svc.stateStore["valid"]; !exists {
		t.Error("有效的 state 不应被清理")
	}
}
