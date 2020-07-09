package main

import (
	"siody.home/om-like/internal/app/test"
	"siody.home/om-like/internal/appmain"
)

func main() {
	appmain.RunApplication("test", test.Bind)
}
