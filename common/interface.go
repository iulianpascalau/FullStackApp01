package common

import "time"

// LoggerFile defines the operations of a component able to write log files
type LoggerFile interface {
	Close() error
	ChangeFileLifeSpan(newDuration time.Duration, sizeInMB uint64) error
}
