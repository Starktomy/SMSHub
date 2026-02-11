package service

import (
	"testing"
	"time"

	"github.com/Starktomy/smshub/internal/models"
)

func TestShouldExecuteTaskNeverRun(t *testing.T) {
	s := &SchedulerService{}
	task := models.ScheduledTask{
		LastRunAt:    0,
		IntervalDays: 7,
		Enabled:      true,
	}

	if !s.shouldExecuteTask(task, time.Now()) {
		t.Error("从未执行过的任务应该执行")
	}
}

func TestShouldExecuteTaskIntervalNotMet(t *testing.T) {
	s := &SchedulerService{}
	now := time.Now()
	task := models.ScheduledTask{
		LastRunAt:     now.Add(-1 * 24 * time.Hour).UnixMilli(), // 1天前执行
		IntervalDays:  7,                                         // 间隔7天
		LastRunStatus: models.LastRunStatusUnknown,
	}

	if s.shouldExecuteTask(task, now) {
		t.Error("间隔未满不应执行")
	}
}

func TestShouldExecuteTaskIntervalMet(t *testing.T) {
	s := &SchedulerService{}
	now := time.Now()
	task := models.ScheduledTask{
		LastRunAt:     now.Add(-8 * 24 * time.Hour).UnixMilli(), // 8天前
		IntervalDays:  7,
		LastRunStatus: models.LastRunStatusUnknown,
	}

	if !s.shouldExecuteTask(task, now) {
		t.Error("间隔已满应该执行")
	}
}

func TestShouldExecuteTaskFailedRetryAfter1Day(t *testing.T) {
	s := &SchedulerService{}
	now := time.Now()

	// 失败且超过1天
	task := models.ScheduledTask{
		LastRunAt:     now.Add(-2 * 24 * time.Hour).UnixMilli(), // 2天前
		IntervalDays:  30,
		LastRunStatus: models.LastRunStatusFailed,
	}

	if !s.shouldExecuteTask(task, now) {
		t.Error("失败任务超过1天应重试")
	}

	// 失败但不到1天
	task2 := models.ScheduledTask{
		LastRunAt:     now.Add(-12 * time.Hour).UnixMilli(), // 12小时前
		IntervalDays:  30,
		LastRunStatus: models.LastRunStatusFailed,
	}

	if s.shouldExecuteTask(task2, now) {
		t.Error("失败任务不到1天不应重试")
	}
}
