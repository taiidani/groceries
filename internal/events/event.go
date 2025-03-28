package events

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Event struct {
	Event string
	Data  fmt.Stringer
}

func (e *Event) String() string {
	return e.Data.String()
}

func (e *Event) Write(w http.ResponseWriter) error {
	if len(e.Event) > 0 {
		fmt.Fprint(w, "event: "+e.Event+"\n")
	}

	if e.Data != nil && e.Data.String() != "" {
		// Place each data line with its own prefix
		// This is to avoid newlines in the data from ending the message early
		datas := strings.Split(e.Data.String(), "\n")
		fmt.Fprint(w, "data: "+strings.Join(datas, "\ndata: ")+"\n")
	} else {
		// Data MUST always be present to trigger events
		fmt.Fprint(w, "data: \n")
	}

	fmt.Fprint(w, "\n")

	f, ok := w.(http.Flusher)
	if !ok {
		return errors.New("client does not support sse")
	}

	f.Flush()
	return nil
}
