package decode

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"k8s.io/klog/v2"
)

// UnGzip 解压缩
func UnGzip(s string) string {
	dByte, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		klog.Error(err)
		return ""
	}
	readData := bytes.NewReader(dByte)
	reader, err := gzip.NewReader(readData)
	if err != nil {
		klog.Error(err)
		return ""
	}
	defer reader.Close()

	res, err := ioutil.ReadAll(readData)
	if err != nil {
		klog.Error(err)
		return ""
	}
	return string(res)

}
