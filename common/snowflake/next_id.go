package main

import (
	"fmt"
	"sync"
	"time"
)

// snowflake雪花算法
// 1. 41位时间戳(毫秒级)
// 2. 10位工作机器ID(分成5位节点ID和5位数据中心ID，最多部署在1024个节点)
// 3. 12位序列号(同一毫秒内最多产生4096个ID)
const (
	workerIDBits     = uint64(5) // 10bit 工作机器ID中的 5bit workerID
	dataCenterIDBits = uint64(5) // 10 bit 工作机器ID中的 5bit dataCenterID
	sequenceBits     = uint64(12)

	maxWorkerID     = int64(-1) ^ (int64(-1) << workerIDBits) //节点ID的最大值 用于防止溢出
	maxDataCenterID = int64(-1) ^ (int64(-1) << dataCenterIDBits)
	maxSequence     = int64(-1) ^ (int64(-1) << sequenceBits)

	timeLeft = uint8(22) // timeLeft = workerIDBits + sequenceBits // 时间戳向左偏移量
	dataLeft = uint8(17) // dataLeft = dataCenterIDBits + sequenceBits
	workLeft = uint8(12) // workLeft = sequenceBits // 节点IDx向左偏移量
	// 2025-1-1 0:00:00 +0800 CST
	twepoch = int64(1735660800000) // 常量时间戳(毫秒) time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local).UnixNano() / 1e6
)

type Worker struct {
	mu           sync.Mutex
	LastStamp    int64 // 记录上一次ID的时间戳
	WorkerID     int64 // 该节点的ID
	DataCenterID int64 // 该节点的 数据中心ID
	Sequence     int64 // 当前毫秒已经生成的ID序列号(从0 开始累加) 1毫秒内最多生成4096个ID
}

// 分布式情况下,我们应通过外部配置文件或其他方式为每台机器分配独立的id
func NewWorker(workerID, dataCenterID int64) *Worker {
	return &Worker{
		WorkerID:     workerID,
		LastStamp:    0,
		Sequence:     0,
		DataCenterID: dataCenterID,
	}
}

func (w *Worker) NextID() uint64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.nextID()
}

func (w *Worker) nextID() uint64 {
	now := w.getMilliSeconds()
	if w.LastStamp == 0 {
		w.LastStamp = now
	}

	if now == w.LastStamp { // 同一毫秒内生成的ID序列号+1
		w.Sequence = (w.Sequence + 1) & maxSequence
		if w.Sequence == 0 { // 序列号溢出，等待下一毫秒
			for now <= w.LastStamp {
				now = w.getMilliSeconds()
			}
		}
	} else { // 不同毫秒内生成的ID序列号置0
		w.Sequence = 0
	}

	w.LastStamp = now

	id := ((now - twepoch) << timeLeft) |
		(w.DataCenterID << dataLeft) |
		(w.WorkerID << workLeft) |
		w.Sequence

	return uint64(id)
}

func (w *Worker) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6 // 纳秒转毫秒
}

func main() {
	start := time.Now()
	fmt.Println("start generate id", start)
	w := NewWorker(1, 1)
	wg := sync.WaitGroup{}
	count := 1000000
	wg.Add(count)
	ch := make(chan uint64)
	for i := 0; i < count; i++ {
		go func() {
			id := w.NextID()
			ch <- id
			wg.Done()
		}()
	}
	m := make(map[uint64]struct{})
	for i := 0; i < count; i++ {
		id := <-ch
		if _, exist := m[id]; exist {
			panic("duplicate id")
		}
		m[id] = struct{}{}
	}
	wg.Wait()
	fmt.Println("All id generated", time.Now(), "cost", time.Since(start))
}
