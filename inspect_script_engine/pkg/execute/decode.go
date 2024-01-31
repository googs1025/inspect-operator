package execute

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"k8s.io/klog/v2"
)

// DecodeBase64 解码 base64
func DecodeBase64(s string) string {
	dByte, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		klog.Error(err)
		return ""
	}
	readData := bytes.NewReader(dByte)
	res, err := ioutil.ReadAll(readData)
	if err != nil {
		klog.Error(err)
		return ""
	}
	return string(res)
}
