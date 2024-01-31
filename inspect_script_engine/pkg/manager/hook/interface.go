package hook

import "scriptimage/pkg/execute"

// WorkflowHook 钩子方法
type WorkflowHook interface {
	BeforeStart(executor *execute.ScriptExecutor) error
	AfterStart(executor *execute.ScriptExecutor) error
}
