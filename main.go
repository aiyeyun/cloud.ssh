package main

import (
	"server"
	"runtime"
)

func main()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
	server.Run()
}
