package execute

import "bytes"

type ScriptExecutor struct {
	// Path 脚本路径
	Path string
	// Script 脚本内容
	Script string
	// TaskName 任务名
	TaskName string
	// Type 类型，分为 local remote
	Type string
	// NodeInfo 宿主机信息
	NodeInfo *Info
	StdOut   bytes.Buffer
	StdErr   bytes.Buffer
}

func NewScriptExecutor(path string, script string, taskName string, t string, nodeInfo *Info) *ScriptExecutor {
	return &ScriptExecutor{
		Path:     path,
		Script:   script,
		TaskName: taskName,
		Type:     t,
		NodeInfo: nodeInfo,
		StdOut:   bytes.Buffer{},
		StdErr:   bytes.Buffer{},
	}
}

type Info struct {
	User     string
	Password string
	Ip       string
}

func NewInfo(user string, password string, ip string) *Info {
	return &Info{User: user, Password: password, Ip: ip}
}
