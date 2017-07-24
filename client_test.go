package httpclient

import (
	"fmt"
	"github.com/golang/glog"
	"testing"
)

func TestGet(t *testing.T) {
	fmt.Println(Get("http://www.baidu.com").Send().String())
}
