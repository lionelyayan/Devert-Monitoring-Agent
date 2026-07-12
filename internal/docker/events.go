package docker

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"

	"github.com/devert/monitor-agent/internal/database"
	"github.com/devert/monitor-agent/internal/webhook"
)

// errorActions adalah kumpulan Docker action yang dianggap error/critical.
// Hanya event dengan action ini yang akan dikirim ke n8n webhook.
var errorActions = map[string]bool{
	"die":     true, // Container mati (exit code != 0 maupun 0)
	"oom":     true, // Container kehabisan memori (Out Of Memory)
	"kill":    true, // Container di-kill paksa
	"stop":    true, // Container dihentikan
	"pause":   true, // Container di-pause
	"destroy": true, // Container dihapus
}

// isErrorEvent mengembalikan true jika event dianggap error/critical.
// Khusus untuk health_status, hanya "unhealthy" yang dikirim.
func isErrorEvent(action, eventType string) bool {
	if eventType == string(events.ContainerEventType) {
		// health_status: unhealthy
		if strings.HasPrefix(action, "health_status") {
			return strings.Contains(action, "unhealthy")
		}
		return errorActions[action]
	}
	return false
}

// DockerEvent is the normalized event structure forwarded to n8n and stored in PostgreSQL.
type DockerEvent struct {
	Server    string    `json:"server"`
	Container string    `json:"container"`
	Image     string    `json:"image"`
	Action    string    `json:"action"`
	EventType string    `json:"event_type"`
	Status    string    `json:"status"`
	Time      time.Time `json:"time"`
}

// EventListener streams Docker events in real-time using the Docker Events API.
// It does NOT poll — it uses a persistent streaming connection.
type EventListener struct {
	client     *client.Client
	serverName string
	db         *database.DB
	webhook    *webhook.Sender
	location   *time.Location
}

// NewEventListener creates an EventListener.
func NewEventListener(cli *client.Client, serverName string, db *database.DB, wh *webhook.Sender, loc *time.Location) *EventListener {
	return &EventListener{
		client:     cli,
		serverName: serverName,
		db:         db,
		webhook:    wh,
		location:   loc,
	}
}

// Listen starts listening to Docker events. It blocks until ctx is cancelled.
// On disconnect it automatically reconnects with exponential backoff.
func (l *EventListener) Listen(ctx context.Context) {
	log.Info().Msg("docker events: starting listener")

	backoff := 1 * time.Second
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("docker events: listener stopped")
			return
		default:
		}

		if err := l.stream(ctx); err != nil && ctx.Err() == nil {
			log.Error().Err(err).Dur("reconnect_in", backoff).Msg("docker events: stream error, reconnecting")
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
		} else {
			backoff = 1 * time.Second
		}
	}
}

// stream opens the Docker Events API stream and processes events until error or ctx cancel.
func (l *EventListener) stream(ctx context.Context) error {
	f := filters.NewArgs()
	// Focus on container, image, volume, and network events
	f.Add("type", string(events.ContainerEventType))
	f.Add("type", string(events.ImageEventType))
	f.Add("type", string(events.VolumeEventType))
	f.Add("type", string(events.NetworkEventType))

	// Docker v27+ uses events.ListOptions instead of types.EventsOptions
	msgCh, errCh := l.client.Events(ctx, events.ListOptions{Filters: f})

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			if err == io.EOF {
				return nil
			}
			return err
		case msg := <-msgCh:
			l.handleEvent(ctx, msg)
		}
	}
}

// handleEvent processes a single Docker event message.
func (l *EventListener) handleEvent(ctx context.Context, msg events.Message) {
	container := msg.Actor.Attributes["name"]
	if container == "" {
		container = msg.Actor.ID
	}

	image := msg.Actor.Attributes["image"]
	if image == "" {
		image = msg.From
	}

	evt := &DockerEvent{
		Server:    l.serverName,
		Container: container,
		Image:     image,
		Action:    string(msg.Action),
		EventType: string(msg.Type),
		Status:    string(msg.Action), // Status mirrors action for Docker events
		Time:      time.Unix(msg.Time, msg.TimeNano/1e9).In(l.location),
	}

	// Hanya proses event yang termasuk kategori error/critical
	if !isErrorEvent(evt.Action, evt.EventType) {
		log.Debug().
			Str("server", evt.Server).
			Str("container", evt.Container).
			Str("action", evt.Action).
			Str("type", evt.EventType).
			Msg("docker event: skipped (non-error)")
		return
	}

	log.Warn().
		Str("server", evt.Server).
		Str("container", evt.Container).
		Str("image", evt.Image).
		Str("action", evt.Action).
		Str("type", evt.EventType).
		Msg("docker event: error/critical detected")

	// Persist to PostgreSQL
	payload, _ := json.Marshal(evt)
	dbEvent := &database.EventLog{
		Server:    evt.Server,
		Container: evt.Container,
		Image:     evt.Image,
		Action:    evt.Action,
		EventType: evt.EventType,
		Status:    evt.Status,
		Message:   evt.EventType + " " + evt.Action + ": " + evt.Container,
		Payload:   payload,
	}
	if err := l.db.InsertEvent(ctx, dbEvent); err != nil {
		log.Error().Err(err).Msg("docker events: failed to persist event")
	}

	// Forward to n8n webhook
	log.Warn().
		Str("server", evt.Server).
		Str("container", evt.Container).
		Str("action", evt.Action).
		Msg("docker events: forwarding to webhook")
	l.webhook.Send(evt)
}
