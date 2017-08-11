package shot

import (
	"fmt"
	"testing"
)

func Test_shot(t *testing.T) {
	defer Release()
	url := "http://127.0.0.1:10005/1050081.html"
	for width := 10; width < 200; width += 10 {
		buf, err := Screenshot(url, width)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(width, len(buf))
	}
}
