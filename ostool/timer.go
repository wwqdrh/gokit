package ostool

import (
	"strconv"
	"time"
)

var TimeDifference int64 = 0

// GetTime get time with rectification
func GetTime() int64 {
	return time.Now().Unix() + TimeDifference
}

// GetTimestamp get current timestamp
func GetTimestamp() string {
	return strconv.FormatInt(GetTime(), 10)
}

// ParseTimestamp parse timestamp to unix time
func ParseTimestamp(timestamp string) int64 {
	unixTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return -1
	}
	return unixTime
}

// FormattedTime get timestamp to print
func FormattedTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
