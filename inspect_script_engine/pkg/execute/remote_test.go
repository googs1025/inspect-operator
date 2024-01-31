package execute

import (
	"log"
	"path"
	"scriptimage/pkg/common"
	"testing"
)

func TestRunRemoteNode(test *testing.T) {

	sc := NewScriptExecutor(
		path.Join(common.GetWd(), common.ScriptFile),
		"bHM=",
		"test",
		"remote",
		NewInfo("root", "googs1025Aa", "1.14.120.233"),
	)

	err := sc.RunRemoteNode()
	if err != nil {
		log.Fatal(err)
	}
}
