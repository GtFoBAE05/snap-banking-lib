package utils

import "time"

func ISO8601Timestamp() string {
	return time.Now().Format("2006-01-02T15:04:05-07:00")
}
