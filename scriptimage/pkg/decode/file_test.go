package decode

import (
	"fmt"
	"testing"
)

func TestGenEncodeFile(test *testing.T) {

	err := WriteStringToFile("H4sIAAAAAAAA/0pNzshXMAQDQAAAAP//B87D+AsAAAA=")
	if err != nil {
		fmt.Println("write err:", err)
		return
	}

	err = GenEncodeFile()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	//execute.RunLocalNode("aaa")
}
