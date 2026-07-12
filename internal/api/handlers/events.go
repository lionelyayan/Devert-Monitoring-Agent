package handlers

import (
	"net/http"
	"strconv"

	"github.com/devert/monitor-agent/internal/database"
)

// EventsHandler handles GET /api/events
type EventsHandler struct {
	db *database.DB
}

func NewEventsHandler(db *database.DB) *EventsHandler {
	return &EventsHandler{db: db}
}

// ListEvents handles GET /api/events with optional query params:
//   - server, container, action, event_type (filter)
//   - limit (default 100, max 500)
//   - offset (default 0)
func (h *EventsHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := database.EventFilter{
		Server:    q.Get("server"),
		Container: q.Get("container"),
		Action:    q.Get("action"),
		EventType: q.Get("event_type"),
		Limit:     limit,
		Offset:    offset,
	}

	events, err := h.db.ListEvents(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if events == nil {
		events = []database.EventLog{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":  len(events),
		"events": events,
	})
}
