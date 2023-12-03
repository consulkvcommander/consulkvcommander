package customcontext

import (
	"context"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
	"sync"
	"time"
)

type contextKey int

const (
	lastInvalidationsOutputKey contextKey = iota
	lastInvalidationsTime
)

type CustomContext struct {
	ctx  context.Context
	lock *sync.RWMutex
}

func New(ctx context.Context) *CustomContext {
	return &CustomContext{ctx: ctx, lock: &sync.RWMutex{}}
}

func (c *CustomContext) GetLastInvalidationsOutput(kvGroupKey string, adaptationMode string) utils.InvalidationsOutput {
	c.lock.RLock()
	defer c.lock.RUnlock()

	output, ok := c.ctx.Value(lastInvalidationsOutputKey).(map[string]utils.InvalidationsOutput)
	if !ok {
		return utils.InvalidationsOutput{}
	}
	return output[kvGroupKey+"-"+adaptationMode]
}

func (c *CustomContext) GetLastInvalidationsTime(kvGroupKey string, adaptationMode string) int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	output, ok := c.ctx.Value(lastInvalidationsTime).(map[string]int64)
	if !ok {
		return 0
	}
	return output[kvGroupKey+"-"+adaptationMode]
}

func (c *CustomContext) SetInvalidationsOutput(kvGroupKey string, newInvalidationsOutput utils.InvalidationsOutput, adaptationMode string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	newInvalidationsOutputMap, ok := c.ctx.Value(lastInvalidationsOutputKey).(map[string]utils.InvalidationsOutput)
	if !ok {
		newInvalidationsOutputMap = map[string]utils.InvalidationsOutput{}
	}
	newInvalidationsTimeMap, ok := c.ctx.Value(lastInvalidationsTime).(map[string]int64)
	if !ok {
		newInvalidationsTimeMap = map[string]int64{}
	}

	newInvalidationsOutputMap[kvGroupKey+"-"+adaptationMode] = newInvalidationsOutput
	newInvalidationsTimeMap[kvGroupKey+"-"+adaptationMode] = time.Now().UnixMilli()

	newCtx := context.WithValue(c.ctx, lastInvalidationsOutputKey, newInvalidationsOutputMap)
	newCtx = context.WithValue(newCtx, lastInvalidationsTime, newInvalidationsTimeMap)

	c.ctx = newCtx
}
