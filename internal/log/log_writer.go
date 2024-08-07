package log

import (
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

type DebugLogWriter struct{}

func (w *DebugLogWriter) Write(p []byte) (n int, err error) {
	log.Debug(string(p))
	return len(p), nil
}

type InfoLogWriter struct{}

func (w *InfoLogWriter) Write(p []byte) (n int, err error) {
	log.Info(string(p))
	return len(p), nil
}

func ShowSpinner(logWriter io.Writer, startStatement, endStatement string) chan struct{} {
	stopSpinnerChan := make(chan struct{})
	go func() {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		for i := 0; ; i++ {
			select {
			case <-stopSpinnerChan:
				logWriter.Write([]byte("\r\033[K"))
				logWriter.Write([]byte(endStatement + "\n"))
				return
			case <-time.After(200 * time.Millisecond):
				logWriter.Write([]byte(fmt.Sprintf("%s %s\r", spinner[i%len(spinner)], startStatement)))
			}
		}
	}()
	return stopSpinnerChan
}
