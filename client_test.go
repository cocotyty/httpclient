package httpclient

import (
	"testing"
	"fmt"
)

func TestGet(t *testing.T) {
	fmt.Println(Get("http://www.baidu.com").Send().String())
}