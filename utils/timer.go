package utils

import (
	"log"
	"runtime"
	"time"
)

func Timer(processName string) func() {
	start := time.Now()
	return func() {
		log.Printf("Process %s took %v\n", processName, time.Since(start))
	}
}

func ShowInfo() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("HeapAlloc = %v mb", toMb(m.HeapAlloc))
	log.Printf("HeapInuse = %v mb", toMb(m.HeapInuse))
	log.Printf("TotalAlloc = %v mb", toMb(m.TotalAlloc))
	log.Printf("Sys = %v mb", toMb(m.Sys))
	log.Printf("NumGC = %v\n", m.NumGC)
}

func toMb(b uint64) uint64 {
	return b / 1024 / 1024
}
