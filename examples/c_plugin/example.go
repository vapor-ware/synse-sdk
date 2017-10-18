package main

// #include "example.h"
import "C"
import "strconv"

func CRead(device int, model string) string {
	reading := C.Read(C.int(device), C.CString(model))
	return strconv.Itoa(int(reading))
}