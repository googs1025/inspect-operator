package log_hook

import (
	"log"
	"scriptimage/pkg/execute"
)

// LoggingHook 日志记录 hook
type LoggingHook struct {
	Logger *log.Logger
}

// BeforeStart 执行前记录日志
func (h *LoggingHook) BeforeStart(executor *execute.ScriptExecutor) error {
	h.Logger.Printf("start to exec script: name: [%v], type: [%v]\n",
		executor.TaskName, executor.Type)
	return nil
}

// AfterStart 执行后记录日志
func (h *LoggingHook) AfterStart(executor *execute.ScriptExecutor) error {
	h.Logger.Printf("completed exec script: stdout: \n%v", executor.StdOut.String())
	h.Logger.Printf("completed exec script: stderr: \n%v", executor.StdErr.String())
	return nil
}
