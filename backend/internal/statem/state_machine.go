// Package statem 提供轻量级业务状态机引擎。
//
// 设计依据：docs/02 ADR-007（状态机实现方案）。
// 核心理念：状态转移以 (From, Action) → (To, Guard, Effect) 显式声明，
// 业务代码统一通过 Engine.Apply() 触发，禁止直接 UPDATE status 字段。
package statem

import (
	"context"
	"fmt"
)

// BizCtx 业务上下文：携带触发者、传入参数、附加元数据。
// 由调用方按需扩展使用 Extra map。
type BizCtx struct {
	Ctx       context.Context
	ActorID   int64
	ActorName string
	ActorRole string
	IP        string
	UA        string
	Payload   map[string]interface{}
	Extra     map[string]interface{}
}

// Transition 单条状态转移定义。
type Transition struct {
	From   string
	Action string
	To     string
	// Guard 守卫函数：返回非 nil 错误则阻断转移。
	Guard func(ctx *BizCtx) error
	// Effect 副作用：在状态成功改变之后调用（如发布事件）。
	Effect func(ctx *BizCtx) error
}

// Engine 状态机引擎。
type Engine struct {
	name        string
	transitions []Transition
}

// New 创建命名状态机。
func New(name string) *Engine {
	return &Engine{name: name, transitions: make([]Transition, 0, 8)}
}

// Define 注册一条状态转移。
func (e *Engine) Define(from, action, to string, guard func(*BizCtx) error, effect func(*BizCtx) error) *Engine {
	e.transitions = append(e.transitions, Transition{
		From:   from,
		Action: action,
		To:     to,
		Guard:  guard,
		Effect: effect,
	})
	return e
}

// Allow 仅注册转移（无守卫/副作用），用于撤回等简单跃迁。
func (e *Engine) Allow(from, action, to string) *Engine {
	return e.Define(from, action, to, nil, nil)
}

// Resolve 在不执行的情况下解析 (from, action) 对应的目标状态。
// 找不到时返回空串与 false。
func (e *Engine) Resolve(from, action string) (string, bool) {
	for _, t := range e.transitions {
		if t.From == from && t.Action == action {
			return t.To, true
		}
	}
	return "", false
}

// Apply 执行状态转移。
// 调用方在 Guard 通过后由自身完成 DB 写入（持久化 from→to）；本引擎仅校验 + 执行 Effect。
// 返回的 to 即新状态；若未匹配到转移，返回错误。
func (e *Engine) Apply(bizCtx *BizCtx, from, action string) (string, error) {
	for _, t := range e.transitions {
		if t.From == from && t.Action == action {
			if t.Guard != nil {
				if err := t.Guard(bizCtx); err != nil {
					return "", err
				}
			}
			// 注意：Effect 的真实持久化由调用方负责；本处仅在 Guard 通过后回调副作用。
			if t.Effect != nil {
				if err := t.Effect(bizCtx); err != nil {
					return "", err
				}
			}
			return t.To, nil
		}
	}
	return "", fmt.Errorf("状态机 %s 不允许从 %s 通过动作 %s 跃迁", e.name, from, action)
}

// Name 返回状态机名称。
func (e *Engine) Name() string { return e.name }
