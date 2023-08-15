package main

import (
	"fmt"
	"testing"
)

func TestIsImage(t *testing.T) {
	inputPath := "E:/workspace/work/compression_tool/image/11.jpg"
	path, name, ty, err := IsImage(inputPath)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Printf("path:%s\nname:%s\nty:%s\n", path, name, ty)
}
