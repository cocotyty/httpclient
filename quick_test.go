package httpclient

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	fmt.Println(Get("http://www.baidu.com").Send().String())
}
