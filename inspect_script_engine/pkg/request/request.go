package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
)

var HttpClient *Http

func init() {
	HttpClient = &Http{
		Client: http.DefaultClient,
	}
}

type Http struct {
	Client *http.Client
}

func (c *Http) DoGet(url string, queryParams map[string]string) (*http.Response, error) {
	klog.Info("get url: ", url, " queryParams: ", queryParams)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	params := req.URL.Query()
	for k, v := range queryParams {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()
	res, err := c.Client.Do(req)
	if err != nil {
		klog.Error("client send err: ", err)
		return nil, err
	}
	return res, nil
}

type Send struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Post 调用消息中心operator接口
func Post(url, title, content string) {

	send := Send{
		Title:   title,
		Content: content,
	}

	body, _ := json.Marshal(send)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body)) // ignore_security_alert

	if err != nil {
		klog.Error("http post request error: ", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			klog.Error("response read error: ", err)
			return
		}
		jsonStr := string(body)
		fmt.Println("Response: ", jsonStr)

	} else {
		klog.Error("Get failed with error: ", resp.Status)
	}

}
