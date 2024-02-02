package v1alpha1

import (
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Inspect
type Inspect struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InspectSpec `json:"spec,omitempty"`

	Status InspectStatus `json:"status,omitempty"`
}

type InspectSpec struct {
	// Type 巡检任务类型: jobs or cronjobs
	Type string `json:"type" default:"jobs"`
	// Schedule 调度时间
	Schedule string `json:"schedule" default:""`
	// GlobalParameters 全局参数
	GlobalParams GlobalParams `json:"globalParams"`
	// Tasks 任务列表
	Tasks []Task `json:"tasks"`
}

// GlobalParams 全局参数
type GlobalParams struct {
	// Env 容器环境变量
	Env []corev1.EnvVar `json:"env,omitempty"`
	// NodeName 选择调度节点
	NodeName string `json:"nodeName,omitempty"`
	// Labels job pod 的 labels
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations job pod 的 annotations
	Annotations map[string]string `json:"annotations,omitempty"`
}

// InspectStatus 任务完成状态
type InspectStatus struct {
	// Status 巡检任务状态
	Status string `json:"status"`
	// 记录执行的 job 状态
	JobResults map[string]v1.JobStatus `json:"jobResults"`
	// 记录执行的 job 状态
	CronJobResults map[string]v1.CronJobStatus `json:"cronJobResults"`
}

// RemoteInfo 登入远端局点需要的信息
type RemoteInfo struct {
	Ip       string `json:"ip"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Task 巡检任务
type Task struct {
	TaskName string `json:"task_name"`
	// Type 任务类型：支持 image or script 方式
	Type string `json:"type"`
	// Source 镜像源：当任务类型为 image 需要填入的镜像
	Source string `json:"source"`
	// Script 脚本内容：当任务类型为 script 填入的内容
	Script string `json:"script"`
	// ScriptLocation 脚本执行地点：本节点或远端节点
	ScriptLocation string `json:"script_location"`
	// RemoteInfos 远端局点信息列表
	RemoteInfos []RemoteInfo `json:"remote_infos"`
}

const (
	Succeed     = "Succeed"     // 代表 JobFlow 中所有 Job 都執行成功
	CronExecute = "CronExecute" // 代表 JobFlow 中所有 Job 都執行成功
	Terminating = "Terminating" // 代表 JobFlow 正在被刪除
	Failed      = "Failed"      // 代表 JobFlow 執行失敗
	Running     = "Running"     // 代表 JobFlow 有任何一個 Job 正在執行
	Pending     = "Pending"     // 代表 JobFlow 正在等待
)

const (
	CronJobsType = "cronjobs" // 代表 JobFlow 有任何一個 Job 正在執行
	JobsType     = "jobs"     // 代表 JobFlow 正在等待
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InspectList
type InspectList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Inspect `json:"items"`
}
