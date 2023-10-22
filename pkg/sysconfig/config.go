package sysconfig

import (
	inspectv1alpha1 "github.com/myoperator/inspectoperator/pkg/apis/inspect/v1alpha1"
	"github.com/myoperator/inspectoperator/pkg/common"
	"io/ioutil"
	"os"
	"sigs.k8s.io/yaml"
)

var SysConfig1 = new(SysConfig)

func InitConfig() error {
	// 读取yaml配置
	config, err := ioutil.ReadFile("./app.yaml")
	if err != nil {
		return err
	}

	//SysConfig = NewSysConfig()

	err = yaml.Unmarshal(config, SysConfig1)
	if err != nil {
		return err
	}

	return nil
}

type Tasks struct {
	Task Task `yaml:"task"`
}

type RemoteIp struct {
	Ip       string `yaml:"ip"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Task struct {
	TaskName       string     `yaml:"task_name"`
	Type           string     `yaml:"type"`
	Source         string     `yaml:"source"`
	ScriptLocation string     `yaml:"script_location"`
	RemoteIps      []RemoteIp `yaml:"remote_ips"`
	Restart        bool       `yaml:"restart"`
}

type SysConfig struct {
	Tasks []Tasks `yaml:"tasks"`
}

func AppConfig(inspect *inspectv1alpha1.Inspect) error {

	// 1. 需要先把SysConfig1中的都删除
	if len(SysConfig1.Tasks) != len(inspect.Spec.Tasks) {
		// 清零后需要先更新app.yaml文件
		SysConfig1.Tasks = make([]Tasks, len(inspect.Spec.Tasks))
		if err := saveConfigToFile(); err != nil {
			return err
		}
	}

	// 2. 更新内存的配置
	for i, task := range inspect.Spec.Tasks {
		SysConfig1.Tasks[i].Task.TaskName = task.Task.TaskName
		SysConfig1.Tasks[i].Task.Type = task.Task.Type
		SysConfig1.Tasks[i].Task.Source = task.Task.Source
		SysConfig1.Tasks[i].Task.ScriptLocation = task.Task.ScriptLocation
		SysConfig1.Tasks[i].Task.Restart = task.Task.Restart
		// 更新远端局点信息
		v := make([]RemoteIp, 0)
		for _, remoteIp := range task.Task.RemoteIps {
			ii := RemoteIp{
				User:     remoteIp.User,
				Password: remoteIp.Password,
				Ip:       remoteIp.Ip,
			}
			v = append(v, ii)

		}
		SysConfig1.Tasks[i].Task.RemoteIps = v
	}
	// 保存配置文件
	if err := saveConfigToFile(); err != nil {
		return err
	}

	return ReloadConfig()
}

// ReloadConfig 重载配置
func ReloadConfig() error {
	return InitConfig()
}

//saveConfigToFile 把config配置放入文件中
func saveConfigToFile() error {

	b, err := yaml.Marshal(SysConfig1)
	if err != nil {
		return err
	}
	// 读取文件
	path := common.GetWd()
	filePath := path + "/app.yaml"
	appYamlFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 644)
	if err != nil {
		return err
	}

	defer appYamlFile.Close()
	_, err = appYamlFile.Write(b)
	if err != nil {
		return err
	}

	return nil
}
