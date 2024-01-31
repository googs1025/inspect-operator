package manager

import (
	"k8s.io/klog/v2"
	"scriptimage/pkg/common"
	"scriptimage/pkg/execute"
	"scriptimage/pkg/manager/hook"
	"sync"
)

// WorkflowManager 管理器
type WorkflowManager struct {
	mu       sync.Mutex
	executor *execute.ScriptExecutor
	beforeFn []func(executor *execute.ScriptExecutor) error
	afterFn  []func(executor *execute.ScriptExecutor) error
}

// NewWorkflowManager 创建
func NewWorkflowManager(sc *execute.ScriptExecutor) *WorkflowManager {
	return &WorkflowManager{
		mu:       sync.Mutex{},
		executor: sc,
	}
}

// RegisterHook 注册资源对象的钩子
func (wm *WorkflowManager) RegisterHook(hook hook.WorkflowHook) {
	wm.beforeFn = append(wm.beforeFn, hook.BeforeStart)
	wm.afterFn = append(wm.afterFn, hook.AfterStart)
}

// Start 执行
func (wm *WorkflowManager) Start() error {

	// 执行创建前的钩子函数
	for _, beforeFn := range wm.beforeFn {
		err := beforeFn(wm.executor)
		if err != nil {
			return err
		}
	}

	err := wm.executor.WriteStringToFile()
	if err != nil {
		klog.Error("write err:", err)
		return err
	}

	err = wm.executor.GenEncodeFile()
	if err != nil {
		klog.Error("write err:", err)
		return err
	}

	if wm.executor.Type == common.RemoteType {
		err = wm.executor.RunRemoteNode()
		if err != nil {
			klog.Error("execute err:", err)
			return err
		}
	} else if wm.executor.Type == common.LocalType {
		err = wm.executor.RunLocalNode()
		if err != nil {
			klog.Error("execute err:", err)
			return err
		}
	}

	// 执行创建后的钩子函数
	for _, afterFn := range wm.afterFn {
		err = afterFn(wm.executor)
		if err != nil {
			return err
		}
	}
	return nil
}
