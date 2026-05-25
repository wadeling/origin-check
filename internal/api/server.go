package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/wadeling/origin-check/internal/store"
)

type Server struct {
	store *store.Store
}

func NewServer(st *store.Store) *Server {
	return &Server{store: st}
}

func (s *Server) Router(corsOrigin string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{corsOrigin, "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
	}))

	r.Get("/health", s.handleHealth)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/relays", s.handleListRelays)
		r.Get("/relays/{id}", s.handleGetRelay)
		r.Get("/relays/{id}/metrics", s.handleRelayMetrics)
		r.Get("/relays/{id}/reports", s.handleRelayReports)
		r.Get("/leaderboard", s.handleLeaderboard)
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListRelays(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.store.ListRelaySummaries(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if summaries == nil {
		summaries = []store.RelaySummary{}
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) handleGetRelay(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	relay, err := s.store.GetRelay(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if relay == nil {
		writeError(w, http.StatusNotFound, errNotFound("relay"))
		return
	}

	since := time.Now().Add(-24 * time.Hour)
	results, err := s.store.GetProbeResults(r.Context(), id, since, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if results == nil {
		results = []store.ProbeResult{}
	}

	avail, _ := s.store.ComputeAvailability24h(r.Context(), id)
	latency, ttft, _ := s.store.GetLatestPerfMetrics(r.Context(), id)

	var authReport *store.AuthenticityReport
	var authScore *float64
	var authVerdict *store.Verdict
	if len(relay.ClaimedModels) > 0 {
		if summary, _ := s.store.GetAuthenticitySummary(r.Context(), id, relay.ClaimedModels); summary != nil {
			authScore = &summary.Score
			authVerdict = &summary.Verdict
		}
		reports, _ := s.store.ListAuthenticityReports(r.Context(), id, relay.ClaimedModels, 1)
		if len(reports) > 0 {
			authReport = &reports[0]
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"relay":                relay,
		"recent_results":       results,
		"availability_24h":     avail,
		"latest_latency_ms":    latency,
		"latest_ttft_ms":       ttft,
		"authenticity_score":   authScore,
		"authenticity_verdict": authVerdict,
		"authenticity_report":  authReport,
	})
}

func (s *Server) handleRelayMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	hours := 24
	points, err := s.store.GetHourlyMetrics(r.Context(), id, hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	aggregates, err := s.store.GetDailyAggregates(r.Context(), id, 7)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"hourly":   points,
		"daily":    aggregates,
	})
}

func (s *Server) handleRelayReports(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	relay, err := s.store.GetRelay(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if relay == nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("relay not found"))
		return
	}

	reports, err := s.store.ListAuthenticityReports(r.Context(), id, relay.ClaimedModels, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if reports == nil {
		reports = []store.AuthenticityReport{}
	}
	writeJSON(w, http.StatusOK, reports)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.store.ListRelaySummaries(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if summaries == nil {
		summaries = []store.RelaySummary{}
	}
	writeJSON(w, http.StatusOK, summaries)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

type notFoundError string

func (e notFoundError) Error() string { return string(e) + " not found" }

func errNotFound(name string) error { return notFoundError(name) }
