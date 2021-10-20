package asciinema

import (
	"encoding/json"
	"fmt"
)

// Header is an Asciicast v2 header line
type Header struct {
	Version   uint              `json:"version"`
	Width     uint              `json:"width"`
	Height    uint              `json:"height"`
	Timestamp int               `json:"timestamp"`
	Command   string            `json:"command"`
	Title     string            `json:"title"`
	Env       map[string]string `json:"env"`
}

// EventType is the type of event (input or output) in an Asciicast v2 line
type EventType string

const (
	// EventTypeOutput is a captured set of bytes sent to the output
	EventTypeOutput EventType = "o"
	// EventTypeInput is a captured input from the user
	EventTypeInput EventType = "i"
)

// Frame is a single line in an Asciicast v2 file
type Frame struct {
	Time      float64
	EventType EventType
	Data      string
}

// MarshalJSON converts a frame into its JSON representation
func (f *Frame) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{f.Time, f.EventType, f.Data})
}

// UnmarshalJSON converts a JSON byte array back into a frame
func (f *Frame) UnmarshalJSON(b []byte) error {
	var rawData []interface{}
	if err := json.Unmarshal(b, &rawData); err != nil {
		return err
	}
	if len(rawData) != 3 {
		return fmt.Errorf("invalid number of fields in Asciicast v2 frame: %v", rawData)
	}
	timestamp, ok := rawData[0].(float64)
	if !ok {
		return fmt.Errorf("the first field in Asciicast v2 frame is not a float: %v", rawData)
	}
	eventType, ok := rawData[1].(string)
	if !ok {
		return fmt.Errorf("the second field in Asciicast v2 frame is not a string: %v", rawData)
	}
	if eventType != string(EventTypeOutput) && eventType != string(EventTypeInput) {
		return fmt.Errorf("the second field in Asciicast v2 frame is not a valid event type: %v", rawData)
	}
	data, ok := rawData[2].(string)
	if !ok {
		return fmt.Errorf("the third field in Asciicast v2 frame is not a string: %v", rawData)
	}
	f.Time = timestamp
	f.EventType = EventType(eventType)
	f.Data = data
	return nil
}
