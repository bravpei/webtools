package utils

import (
	"fmt"
	"hash/fnv"
	"log/slog"
)

// DispatcherTask 包含需要执行的函数和数据
type DispatcherTask struct {
	Key  string                          // 任务的标识键
	Data interface{}                     // 任务数据
	F    func(string, interface{}) error // 处理函数
}

// Dispatcher 用于分发任务到对应的协程
type Dispatcher struct {
	name      string
	shards    []chan DispatcherTask
	numShards uint32
}

// NewDispatcher 创建并初始化Dispatcher
func NewDispatcher(name string, numShards, size uint32) *Dispatcher {
	d := &Dispatcher{
		shards:    make([]chan DispatcherTask, numShards),
		numShards: numShards,
		name:      name,
	}
	// 初始化每个分片的通道，并启动处理协程
	for i := uint32(0); i < numShards; i++ {
		ch := make(chan DispatcherTask, size) // 带缓冲的通道
		d.shards[i] = ch
		go processTasks(ch)
	}
	return d
}

// processTasks 处理任务
func processTasks(ch <-chan DispatcherTask) {
	for task := range ch {
		err := task.F(task.Key, task.Data)
		if err != nil {
			slog.Error("处理任务失败", "key", task.Key, "error", err)
			continue
		}
	}
}

// Dispatch 根据Key分发任务到对应的分片
func (d *Dispatcher) Dispatch(t DispatcherTask) (err error) {
	shard := hash(t.Key) % d.numShards
	select {
	case d.shards[shard] <- t:
		slog.Info("提交任务到分片成功", "key", t.Key, "name", d.name, "shard", shard, "channelLength", len(d.shards[shard]))
	default:
		err = fmt.Errorf("%s:提交%s任务到分片-%d:失败,chan长度:%d", t.Key, d.name, shard, len(d.shards[shard]))
	}
	return
}

// hash 计算字符串的哈希值（使用FNV-1a）
func hash(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		slog.Error("计算哈希值失败", "error", err)
	}
	return h.Sum32()
}
