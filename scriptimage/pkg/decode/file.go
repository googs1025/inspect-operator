package decode

import (
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
	"scriptimage/pkg/common"
)

// GenEncodeFile 生成脚本
func GenEncodeFile() error {
	path := common.GetWd()
	f, err := os.OpenFile(path+common.ScriptFile, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	decode := UnGzip(string(b)) //反解 之后的 字符串 ,重新写入
	err = f.Truncate(0)         //清空文件1
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0) //清空文件2
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(decode))
	if err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}

// WriteStringToFile 把脚本内容写入文件
func WriteStringToFile(script string) error {
	path := common.GetWd()
	dstFile, err := os.Create(path + common.ScriptFile)
	if err != nil {
		klog.Error(err.Error())
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.WriteString(script + "\n")
	if err != nil {
		klog.Error(err.Error())
		return err
	}
	return nil
}
