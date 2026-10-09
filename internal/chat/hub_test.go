package chat

import (
	"sync"
	"testing"
	"time"
)

func TestChatHubMultiplayer(t *testing.T) {
	hub := NewHub(10)

	// Register 3 clients
	c1 := hub.Register("sess_1", "alice", "testing", "Lobby")
	c2 := hub.Register("sess_2", "bob", "lurking", "Lobby")
	c3 := hub.Register("sess_3", "charlie", "wandering", "Archive")

	// Allow time for registration goroutines
	time.Sleep(50 * time.Millisecond)

	if hub.OnlineCount() != 3 {
		t.Fatalf("Expected 3 online users, got %d", hub.OnlineCount())
	}

	dist := hub.RoomDistribution()
	if dist["Lobby"] != 2 || dist["Archive"] != 1 {
		t.Errorf("Unexpected room distribution: %+v", dist)
	}

	// Test broadcast
	hub.BroadcastUser("alice", "Hello the wire!", "#00e5ff")

	// Check alice, bob, and charlie receive message
	received := sync.WaitGroup{}
	received.Add(3)

	checkClient := func(c *Client) {
		defer received.Done()
		timeout := time.After(500 * time.Millisecond)
		for {
			select {
			case msg := <-c.SendCh:
				if msg.Content == "Hello the wire!" {
					return
				}
			case <-timeout:
				t.Errorf("Client %s timed out waiting for broadcast", c.Nickname)
				return
			}
		}
	}

	go checkClient(c1)
	go checkClient(c2)
	go checkClient(c3)

	received.Wait()

	// Test mute functionality
	hub.SetMute("sess_2", true)
	hub.BroadcastUser("alice", "Can bob hear this?", "#00e5ff")

	time.Sleep(50 * time.Millisecond)
	select {
	case msg := <-c2.SendCh:
		if msg.Content == "Can bob hear this?" {
			t.Errorf("Muted client c2 should not have received user message")
		}
	default:
		// Expected: nothing received while muted
	}

	// Test disconnect
	hub.Unregister("sess_3")
	time.Sleep(50 * time.Millisecond)

	if hub.OnlineCount() != 2 {
		t.Errorf("Expected 2 online users after disconnect, got %d", hub.OnlineCount())
	}
}

func TestChatHubHistoryBounds(t *testing.T) {
	hub := NewHub(5)

	for i := 0; i < 10; i++ {
		hub.BroadcastUser("system", "msg", "")
	}

	time.Sleep(50 * time.Millisecond)
	history := hub.GetRecentHistory()
	if len(history) > 5 {
		t.Errorf("History length %d exceeds max bounded limit 5", len(history))
	}
}
