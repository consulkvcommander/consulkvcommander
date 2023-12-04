package knowledgebase

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

type KnowledgeBaseContext struct {
	ctx  context.Context
	lock *sync.RWMutex
}

func New(ctx context.Context) *KnowledgeBaseContext {
	return &KnowledgeBaseContext{ctx: ctx, lock: &sync.RWMutex{}}
}

func (c *KnowledgeBaseContext) GetLastInvalidationsOutput(consulKvKey string, adaptationMode string) utils.InvalidationsOutput {
	c.lock.RLock()
	defer c.lock.RUnlock()

	output, ok := c.ctx.Value(lastInvalidationsOutputKey).(map[string]utils.InvalidationsOutput)
	if !ok {
		return utils.InvalidationsOutput{}
	}
	return output[consulKvKey+"-"+adaptationMode]
}

func (c *KnowledgeBaseContext) GetLastInvalidationsTime(consulKvKey string, adaptationMode string) int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	output, ok := c.ctx.Value(lastInvalidationsTime).(map[string]int64)
	if !ok {
		return 0
	}
	return output[consulKvKey+"-"+adaptationMode]
}

func (c *KnowledgeBaseContext) SetInvalidationsOutput(consulKvKey string, newInvalidationsOutput utils.InvalidationsOutput, adaptationMode string) {
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

	newInvalidationsOutputMap[consulKvKey+"-"+adaptationMode] = newInvalidationsOutput
	newInvalidationsTimeMap[consulKvKey+"-"+adaptationMode] = time.Now().UnixMilli()

	newCtx := context.WithValue(c.ctx, lastInvalidationsOutputKey, newInvalidationsOutputMap)
	newCtx = context.WithValue(newCtx, lastInvalidationsTime, newInvalidationsTimeMap)

	c.ctx = newCtx
}
