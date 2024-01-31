package execute

import (
	"log"
	"path"
	"scriptimage/pkg/common"
	"testing"
)

func TestRunLocalNode(test *testing.T) {

	sc := NewScriptExecutor(
		path.Join(common.GetWd(), common.ScriptFile),
		"bHM=",
		"test",
		"local",
		NewInfo("", "", ""),
	)

	err := sc.RunRemoteNode()
	if err != nil {
		log.Fatal(err)
	}
}
