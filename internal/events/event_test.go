package events

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testStringer struct {
	value string
}

func (t testStringer) String() string {
	return t.value
}

func TestEvent_String(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  string
	}{
		{
			name: "simple string data",
			event: Event{
				Event: "update",
				Data:  testStringer{value: "test data"},
			},
			want: "test data",
		},
		{
			name: "empty data",
			event: Event{
				Event: "ping",
				Data:  testStringer{value: ""},
			},
			want: "",
		},
		{
			name: "multiline data",
			event: Event{
				Event: "message",
				Data:  testStringer{value: "line1\nline2\nline3"},
			},
			want: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.String()
			if got != tt.want {
				t.Errorf("Event.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_Write(t *testing.T) {
	tests := []struct {
		name       string
		event      Event
		wantOutput string
		wantErr    bool
	}{
		{
			name: "event with simple data",
			event: Event{
				Event: "update",
				Data:  testStringer{value: "test data"},
			},
			wantOutput: "event: update\ndata: test data\n\n",
			wantErr:    false,
		},
		{
			name: "event with empty name",
			event: Event{
				Event: "",
				Data:  testStringer{value: "data only"},
			},
			wantOutput: "data: data only\n\n",
			wantErr:    false,
		},
		{
			name: "event with nil data",
			event: Event{
				Event: "ping",
				Data:  nil,
			},
			wantOutput: "event: ping\ndata: \n\n",
			wantErr:    false,
		},
		{
			name: "event with empty data",
			event: Event{
				Event: "empty",
				Data:  testStringer{value: ""},
			},
			wantOutput: "event: empty\ndata: \n\n",
			wantErr:    false,
		},
		{
			name: "event with multiline data",
			event: Event{
				Event: "message",
				Data:  testStringer{value: "line1\nline2\nline3"},
			},
			wantOutput: "event: message\ndata: line1\ndata: line2\ndata: line3\n\n",
			wantErr:    false,
		},
		{
			name: "event with trailing newline in data",
			event: Event{
				Event: "test",
				Data:  testStringer{value: "data\n"},
			},
			wantOutput: "event: test\ndata: data\ndata: \n\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			err := tt.event.Write(recorder)

			if (err != nil) != tt.wantErr {
				t.Errorf("Event.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := recorder.Body.String()
			if got != tt.wantOutput {
				t.Errorf("Event.Write() output = %q, want %q", got, tt.wantOutput)
			}
		})
	}
}

func TestEvent_Write_Flusher(t *testing.T) {
	// Test that Write() calls Flush() on a ResponseWriter that supports flushing
	event := Event{
		Event: "test",
		Data:  testStringer{value: "data"},
	}

	recorder := httptest.NewRecorder()
	err := event.Write(recorder)

	if err != nil {
		t.Errorf("Event.Write() unexpected error = %v", err)
	}

	// httptest.ResponseRecorder supports Flusher interface
	if !recorder.Flushed {
		t.Error("Event.Write() did not flush the response")
	}
}

type nonFlusher struct {
	http.ResponseWriter
	written strings.Builder
}

func (n *nonFlusher) Write(p []byte) (int, error) {
	return n.written.Write(p)
}

func (n *nonFlusher) Header() http.Header {
	return http.Header{}
}

func (n *nonFlusher) WriteHeader(statusCode int) {}

func TestEvent_Write_NonFlusher(t *testing.T) {
	// Test that Write() returns an error when ResponseWriter doesn't support Flusher
	event := Event{
		Event: "test",
		Data:  testStringer{value: "data"},
	}

	w := &nonFlusher{}
	err := event.Write(w)

	if err == nil {
		t.Error("Event.Write() expected error for non-flusher writer, got nil")
	}

	expectedErr := "client does not support sse"
	if err.Error() != expectedErr {
		t.Errorf("Event.Write() error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestEvent_Write_ComplexData(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "data with special characters",
			data: "test: value\nwith: colons",
		},
		{
			name: "data with multiple consecutive newlines",
			data: "line1\n\n\nline2",
		},
		{
			name: "data with only newlines",
			data: "\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := Event{
				Event: "complex",
				Data:  testStringer{value: tt.data},
			}

			recorder := httptest.NewRecorder()
			err := event.Write(recorder)

			if err != nil {
				t.Errorf("Event.Write() unexpected error = %v", err)
			}

			output := recorder.Body.String()

			// Verify that output starts with event name
			if !strings.HasPrefix(output, "event: complex\n") {
				t.Error("Output should start with event name")
			}

			// Verify that output has data prefix
			if !strings.Contains(output, "data: ") {
				t.Error("Output should contain data prefix")
			}

			// Verify that output ends with double newline
			if !strings.HasSuffix(output, "\n\n") {
				t.Error("Output should end with double newline")
			}

			// Count the number of "data: " prefixes - should match newlines + 1
			expectedDataPrefixes := strings.Count(tt.data, "\n") + 1
			actualDataPrefixes := strings.Count(output, "data: ")
			if actualDataPrefixes != expectedDataPrefixes {
				t.Errorf("Expected %d data prefixes, got %d", expectedDataPrefixes, actualDataPrefixes)
			}
		})
	}
}

func ExampleEvent_Write() {
	event := Event{
		Event: "update",
		Data:  testStringer{value: "Hello, World!"},
	}

	recorder := httptest.NewRecorder()
	err := event.Write(recorder)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Print(recorder.Body.String())
	// Output:
	// event: update
	// data: Hello, World!
	//
}
