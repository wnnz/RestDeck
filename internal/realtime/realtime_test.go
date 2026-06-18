package realtime

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"restdeck/internal/domain"
)

func TestWebSocketEcho(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		defer conn.Close()
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, append([]byte("echo:"), data...)); err != nil {
			t.Fatalf("write: %v", err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	result := NewService().TestWebSocket(t.Context(), WebSocketRequest{URL: wsURL, Message: "hello", TimeoutMs: 3000}, emptyEnv(), nil)
	if result.Error != "" {
		t.Fatalf("websocket error: %s", result.Error)
	}
	if len(result.Received) != 1 || result.Received[0] != "echo:hello" {
		t.Fatalf("unexpected websocket result: %#v", result)
	}
}

func TestSSECollectsEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for i := 1; i <= 2; i++ {
			_, _ = fmt.Fprintf(w, "event: message\ndata: %d\n\n", i)
		}
	}))
	defer server.Close()

	result := NewService().TestSSE(t.Context(), SSERequest{URL: server.URL, TimeoutMs: 3000, MaxEvents: 2}, emptyEnv(), nil)
	if result.Error != "" {
		t.Fatalf("sse error: %s", result.Error)
	}
	if len(result.Events) != 2 {
		t.Fatalf("events = %#v", result.Events)
	}
}

func emptyEnv() domain.Environment {
	return domain.Environment{}
}
