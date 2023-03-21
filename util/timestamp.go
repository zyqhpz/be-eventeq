package util

import (
	"time"
)

/*
* Get current timestamp in GMT +8
 */
func GetCurrentTime() time.Time {
	return time.Now().Add(8 * time.Hour)
}