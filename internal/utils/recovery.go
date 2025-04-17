package utils

import (
	"sync"
)

func Recovery(wg *sync.WaitGroup) {
	if r := recover(); r != nil {
		wg.Done()
	}
}

func RecoveryWithCallback(wg *sync.WaitGroup, callback func(any)) {
	if r := recover(); r != nil {
		if callback != nil {
			callback(r)
		}
		wg.Done()
	}
}
