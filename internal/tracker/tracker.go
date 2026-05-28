package tracker

import (
	"sync"
	"time"
)

type Position struct {
	ID        string    `json:"id"`
	Token     string    `json:"token,omitempty"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	TargetLat float64   `json:"target_lat"`
	TargetLon float64   `json:"target_lon"`
	Status    string    `json:"status,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Step      int       `json:"step"`

}

type Tracker struct {
	mu        sync.RWMutex
	positions map[string]Position
	clients   map[chan Position]struct{}
}

func NewTracker() *Tracker {
	return &Tracker{
		positions: make(map[string]Position),
		clients:   make(map[chan Position]struct{}),
	}
}

func (t *Tracker) GetPositions() []Position {
	t.mu.RLock()
	defer t.mu.RUnlock()

	positions := make([]Position, 0, len(t.positions))
	for _, p := range t.positions {
		positions = append(positions, p)
	}
	return positions
}

func (t *Tracker) GetPosition(id string) (Position, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	p, ok := t.positions[id]
	return p, ok
}

func (t *Tracker) SetPosition(p Position) {
	t.mu.Lock()
	t.positions[p.ID] = p

	clients := make([]chan Position, 0, len(t.clients))
	for ch := range t.clients {
		clients = append(clients, ch)
	}
	t.mu.Unlock()

	for _, ch := range clients {
		select {
		case ch <- p:
		default:
			// cliente lento, ignora este envio
		}
	}
}

func (t *Tracker) Subscribe() chan Position {
	ch := make(chan Position, 8)

	t.mu.Lock()
	t.clients[ch] = struct{}{}
	t.mu.Unlock()

	return ch
}

func (t *Tracker) Unsubscribe(ch chan Position) {
	t.mu.Lock()
	delete(t.clients, ch)
	t.mu.Unlock()
	close(ch)
}
