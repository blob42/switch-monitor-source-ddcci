// Package main provides ...
package main

import (
	"fmt"
	"syscall"
	"unsafe"
	//"io/ioutil"
	 //"github.com/fvbock/endless"
	 "github.com/gin-gonic/gin"
	//"log"
	 "net/http"
	// "os"
	// "os/exec"
)

const (
	MOUSE_X = 1000
	MOUSE_Y = 1000
	M_POINT = ( MOUSE_X & 0xFFFFFFFF ) | ( MOUSE_Y << 32 )
	VPC_SOURCE_SELECT = '\x60'
	WINDOWS_M = 15
	LINUX_M = 17
	
	WINDOWS_PARAM = "windows"
	LINUX_PARAM = "linux"
)

var (
	user32, _ = syscall.LoadLibrary("User32.dll")
	dxva2, _ = syscall.LoadLibrary("dxva2.dll")
	monitorFromPoint, _ = syscall.GetProcAddress(user32, "MonitorFromPoint")
	GetPhysicalMonitorsFromHMONITOR, _ = syscall.GetProcAddress(dxva2, "GetPhysicalMonitorsFromHMONITOR")
	SetVCPFeature, _ = syscall.GetProcAddress(dxva2, "SetVCPFeature")
	DestroyPhysicalMonitor, _  = syscall.GetProcAddress(dxva2, "DestroyPhysicalMonitor")
)

 func switchMonitorHandler(c *gin.Context) {
 
	os := c.Param("os")

	switch os {
		case WINDOWS_PARAM:
			go setMonitorInputSource(WINDOWS_M)
		case LINUX_PARAM:
			go setMonitorInputSource(LINUX_M)
	}

	//fmt.Printf("Received command %s", command)
	c.String(http.StatusOK, fmt.Sprintf("monitor switched to %s!", os))
} 


func getMonitorHandle() (result uintptr) {
		var nargs uintptr = 2
		ret, _, callErr := syscall.Syscall(
			uintptr(monitorFromPoint),
			nargs,
			uintptr(M_POINT),
			uintptr(1),
			0)
		
		if callErr != 0 {
			abort("Call getMonitorHandle", callErr)	
		}
		
		res := uintptr(ret)
		result = getPhysicalMonitor(res)
		return
}

func getPhysicalMonitor(handle uintptr) (result uintptr) {
	b := make([]byte, 256)
	var nargs uintptr = 3
	_, _, callErr := syscall.Syscall(
		uintptr(GetPhysicalMonitorsFromHMONITOR),
		nargs,
		handle,
		uintptr(1),
		uintptr(unsafe.Pointer(&b[0])))
	
	if callErr != 0 {
			abort("Call getPhysicalMonitor", callErr)
	}
	result = uintptr(b[0])
	return
}

func setMonitorInputSource(source int) {
	mHandle := getMonitorHandle()
	
	var nargs uintptr = 3
	_, _, callErr := syscall.Syscall(
		uintptr(SetVCPFeature),
		nargs,
		uintptr(mHandle),
		VPC_SOURCE_SELECT,
		uintptr(source))
		
	if callErr != 0 {
			abort("Call SetVCPFeature", callErr)
	}
	destroyPhysicalMonitor(mHandle)
}

func destroyPhysicalMonitor(monitor uintptr) {
	var nargs uintptr = 1
	_, _, callErr := syscall.Syscall(
		uintptr(DestroyPhysicalMonitor),
		nargs,
		monitor,
		0,
		0)
		
	if callErr != 0 {
			abort("Call destroyPhysicalMonitor", callErr)
	}
}

func abort(funcname string, err error) {
    panic(fmt.Sprintf("%s failed: %v", funcname, err))
}



func main() {
	defer syscall.FreeLibrary(user32)
	defer syscall.FreeLibrary(dxva2)

	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	monitor := router.Group("/monitor")
	{
		monitor.GET("/switch/:os", switchMonitorHandler)
	}

	router.Run(":7500")
	//endless.ListenAndServe(":7500", router)
}
