package asciinema

import "encoding/json"

type AsciicastHeader struct {
	Version   uint              `json:"version"`
	Width     uint              `json:"width"`
	Height    uint              `json:"height"`
	Timestamp int               `json:"timestamp"`
	Command   string            `json:"command"`
	Title     string            `json:"title"`
	Env       map[string]string `json:"env"`
}

type AsciicastEventType string

const (
	AsciicastEventTypeOutput AsciicastEventType = "o"
	//AsciicastEventTypeInput AsciicastEventType = "i"
)

type AsciicastFrame struct {
	Time      float64
	EventType AsciicastEventType
	Data      string
}

func (f *AsciicastFrame) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{f.Time, f.EventType, f.Data})
}
