package data

import "github.com/google/uuid"

// HeapdumpJob is what is stored in the queue of heapdumps to
// be analysed
type HeapdumpJob struct {
	Id           string `json:"id"`
	HprofFile    string `json:"zipfile"`
	QueuedMillis int64  `json:"queued_millis"`
}

func NewHeapdumpJob(hprofFile string) *HeapdumpJob {
	return &HeapdumpJob{
		Id:        uuid.New().String(),
		HprofFile: hprofFile,
	}
}
