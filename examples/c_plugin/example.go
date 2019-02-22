package main

// #include "example.h"
import "C"
import "strconv"

func cRead(device int, dtype string) string {
	reading := C.Read(C.int(device), C.CString(dtype))
	return strconv.Itoa(int(reading))
}
