package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/containerssh/containerssh/audit/format"
	"log"
	"os"
)

func main() {
	file := ""
	flag.StringVar(&file, "file", "", "File to process")
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}

	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	fh, err := os.Open(file)
	if err != nil {
		log.Fatalf("failed to open audit log file %s (%v)", file, err)
	}

	messages, errors, done := format.Decode(fh)
	for {
		var msg *format.DecodedMessage
		select {
		case msg = <-messages:
			if msg == nil {
				break
			}

			var data []byte
			data, err = json.Marshal(msg)
			if err != nil {
				structuredError := map[string]string{
					"error": fmt.Sprintf("JSON encoding error: (%v)", err),
				}
				data, _ = json.Marshal(structuredError)
				_, _ = os.Stdout.Write(data)
				_, _ = os.Stdout.Write([]byte("\n"))
			} else if data != nil {
				_, _ = os.Stdout.Write(data)
				_, _ = os.Stdout.Write([]byte("\n"))
			} else {
				break
			}
		case channelError := <-errors:
			if channelError != nil {
				structuredError := map[string]string{
					"error": channelError.Error(),
				}
				data, _ := json.Marshal(structuredError)
				_, _ = os.Stdout.Write(data)
				_, _ = os.Stdout.Write([]byte("\n"))
			} else {
				break
			}
		case <-done:
			return
		}
	}
}
