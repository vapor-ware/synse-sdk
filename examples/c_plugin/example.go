package main

// #include "example.h"
import "C"


func Read(device int, model string) string {
	reading := C.Read(C.int(device), C.CString(model))
	return reading
}