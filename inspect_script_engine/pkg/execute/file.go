package execute

import (
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
)

// GenEncodeFile 生成脚本
func (sc *ScriptExecutor) GenEncodeFile() error {
	f, err := os.OpenFile(sc.Path, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	//反解 之后的 字符串 ,重新写入
	decode := DecodeBase64(string(b))
	err = f.Truncate(0) //清空文件1
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
func (sc *ScriptExecutor) WriteStringToFile() error {
	dstFile, err := os.Create(sc.Path)
	if err != nil {
		klog.Error(err.Error())
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.WriteString(sc.Script + "\n")
	if err != nil {
		klog.Error(err.Error())
		return err
	}
	return nil
}
