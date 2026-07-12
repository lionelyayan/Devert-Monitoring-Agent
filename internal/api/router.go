package api

import (
	"net/http"
	"time"

	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/devert/monitor-agent/internal/api/handlers"
	"github.com/devert/monitor-agent/internal/api/middleware"
	"github.com/devert/monitor-agent/internal/database"
)

// NewRouter builds and returns the chi router with all routes registered.
func NewRouter(
	apiToken string,
	rateLimitRPM int,
	serverName string,
	cli *client.Client,
	db *database.DB,
) http.Handler {
	r := chi.NewRouter()

	// ── Global middlewares ─────────────────────────────────────────
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)
	r.Use(middleware.RateLimiter(rateLimitRPM))

	// ── Health check (unauthenticated) ────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"devert-monitor-agent"}`))
	})

	// ── Handlers ──────────────────────────────────────────────────
	serverH := handlers.NewServerHandler(serverName)
	containersH := handlers.NewContainersHandler(cli)
	imagesH := handlers.NewImagesHandler(cli)
	networksH := handlers.NewNetworksHandler(cli)
	volumesH := handlers.NewVolumesHandler(cli)
	servicesH := handlers.NewServicesHandler()
	resourcesH := handlers.NewResourcesHandler(cli)
	eventsH := handlers.NewEventsHandler(db)

	// ── Protected API routes ──────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.BearerAuth(apiToken))

		// Server
		r.Get("/api/server", serverH.GetServer)

		// Containers
		r.Get("/api/containers", containersH.ListContainers)
		r.Get("/api/container/{id}", containersH.GetContainer)

		// Container actions
		r.Post("/api/container/{id}/start", containersH.StartContainer)
		r.Post("/api/container/{id}/stop", containersH.StopContainer)
		r.Post("/api/container/{id}/restart", containersH.RestartContainer)
		r.Post("/api/container/{id}/remove", containersH.RemoveContainer)

		// Images
		r.Get("/api/images", imagesH.ListImages)

		// Networks
		r.Get("/api/networks", networksH.ListNetworks)

		// Volumes
		r.Get("/api/volumes", volumesH.ListVolumes)

		// Services
		r.Get("/api/services", servicesH.ListServices)

		// Container resources
		r.Get("/api/resources", resourcesH.ListResources)

		// Events log
		r.Get("/api/events", eventsH.ListEvents)
	})

	return r
}

// corsMiddleware adds permissive CORS headers. Tighten in production.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
