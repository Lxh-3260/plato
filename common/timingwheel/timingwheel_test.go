package timingwheel

import (
	"testing"
	"time"
)

func TestTimingWheel_AfterFunc(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	durations := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}
	for _, d := range durations {
		t.Run("", func(t *testing.T) {
			exitC := make(chan time.Time) // channel 在定时器触发时传递时间

			start := time.Now().UTC()
			tw.AfterFunc(d, func() {
				exitC <- time.Now().UTC()
			}) // 使用定时轮（timing wheel）创建一个定时器，定时器在 d 时间后触发，触发时调用回调函数 func()将当前时间写入 exitC的channel

			got := (<-exitC).Truncate(time.Millisecond)    // 截断到ms
			min := start.Add(d).Truncate(time.Millisecond) // 开始时间加上d时间，截断到ms

			err := 5 * time.Millisecond // 误差允许5ms以内
			if got.Before(min) || got.After(min.Add(err)) {
				t.Errorf("Timer(%s) expiration: want [%s, %s], got %s", d, min, min.Add(err), got)
			}
		})
	}
	/*
		cmd: go test -v -bench=. -benchmem
		参数解读:
		1. -bench=. 表示运行所有的基准测试
		如果想要只运行某个基准测试，可以使用-bench=后面跟基准测试的名字（函数名为Test后面的，例如本函数为TimingWheel_AfterFunc）
		2. -benchmem 表示在基准测试结束后输出内存分配的统计数据
		output:
		=== RUN   TestTimingWheel_AfterFunc
		=== RUN   TestTimingWheel_AfterFunc/#00
		=== RUN   TestTimingWheel_AfterFunc/#01
		=== RUN   TestTimingWheel_AfterFunc/#02
		=== RUN   TestTimingWheel_AfterFunc/#03
		=== RUN   TestTimingWheel_AfterFunc/#04
		=== RUN   TestTimingWheel_AfterFunc/#05
		=== RUN   TestTimingWheel_AfterFunc/#06
		--- PASS: TestTimingWheel_AfterFunc (1.67s)
			--- PASS: TestTimingWheel_AfterFunc/#00 (0.00s)
			--- PASS: TestTimingWheel_AfterFunc/#01 (0.01s)
			--- PASS: TestTimingWheel_AfterFunc/#02 (0.01s)
			--- PASS: TestTimingWheel_AfterFunc/#03 (0.05s)
			--- PASS: TestTimingWheel_AfterFunc/#04 (0.10s)
			--- PASS: TestTimingWheel_AfterFunc/#05 (0.50s)
			--- PASS: TestTimingWheel_AfterFunc/#06 (1.00s)
	*/
}

type scheduler struct {
	intervals []time.Duration
	current   int
}

func (s *scheduler) Next(prev time.Time) time.Time {
	if s.current >= len(s.intervals) {
		return time.Time{}
	}
	next := prev.Add(s.intervals[s.current])
	s.current += 1
	return next
}

func TestTimingWheel_ScheduleFunc(t *testing.T) {
	tw := NewTimingWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	s := &scheduler{intervals: []time.Duration{
		1 * time.Millisecond,   // start + 1ms
		4 * time.Millisecond,   // start + 5ms
		5 * time.Millisecond,   // start + 10ms
		40 * time.Millisecond,  // start + 50ms
		50 * time.Millisecond,  // start + 100ms
		400 * time.Millisecond, // start + 500ms
		500 * time.Millisecond, // start + 1s
	}}

	exitC := make(chan time.Time, len(s.intervals))

	start := time.Now().UTC()
	tw.ScheduleFunc(s, func() {
		exitC <- time.Now().UTC()
	})

	accum := time.Duration(0)
	for _, d := range s.intervals {
		got := (<-exitC).Truncate(time.Millisecond)
		accum += d
		min := start.Add(accum).Truncate(time.Millisecond)

		err := 5 * time.Millisecond
		if got.Before(min) || got.After(min.Add(err)) {
			t.Errorf("Timer(%s) expiration: want [%s, %s], got %s", accum, min, min.Add(err), got)
		}
	}
}
