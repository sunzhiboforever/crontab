package worker

import (
	"github.com/sunzhiboforever/crontab/common"
	"github.com/sunzhiboforever/crontab/drivers"
	"time"
)

type Schedule struct {
	// 任务列表
	jobs map[string]*drivers.Job
	// 事件通道
	eventChan chan *drivers.Event
	// 有新的事件来的时候，要通知调度器重新计算执行时间
	eventNotifyChan chan int
}

// 初始化调度器
func InitSchedule(kv drivers.KvWorker) (err error) {
	var (
		schedule *Schedule
	)
	schedule = &Schedule{}
	// 开启监听，获取事件到监听通道
	schedule.eventChan = make(chan *drivers.Event, 1000)
	schedule.eventChan, err = kv.Watch()
	if err != nil {
		return
	}
	schedule.jobs = make(map[string]*drivers.Job)
	schedule.eventNotifyChan = make(chan int, 0)

	// 启动事件处理
	go schedule.event()

	// 启动任务处理
	go schedule.schedule()

	return
}

// 事件处理器
func (s *Schedule) event() {
	var (
		event *drivers.Event
	)
	for event = range s.eventChan {
		switch event.Type {
		case common.EVENT_TYPE_SAVE:
			s.eventSave(event)
		case common.EVENT_TYPE_DELETE:
			s.eventDelete(event)
		}
	}
}

// 处理保存事件
func (s *Schedule) eventSave(event *drivers.Event) {
	s.jobs[event.Name] = event.Job
}

// 处理删除事件
func (s *Schedule) eventDelete(event *drivers.Event) {
	if _, ok := s.jobs[event.Name]; ok {
		delete(s.jobs, event.Name)
	}
}

// 调度器
func (s *Schedule) schedule() {
	var (
		nearbyTime time.Duration
		timer *time.Timer
	)
	timer = time.NewTimer(nearbyTime)
	for {
		select {
		case <-s.eventNotifyChan:
			// 重新计算一遍
		case <-timer.C:

		}
	}
}
