package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wadeling/origin-check/internal/analyzer"
	"github.com/wadeling/origin-check/internal/crypto"
	"github.com/wadeling/origin-check/internal/probe"
	"github.com/wadeling/origin-check/internal/queue"
	"github.com/wadeling/origin-check/internal/store"
)

type Worker struct {
	store    *store.Store
	queue    *queue.Queue
	probe    *probe.Client
	analyzer *analyzer.Engine
	enc      *crypto.Encryptor
}

func New(st *store.Store, q *queue.Queue, enc *crypto.Encryptor) *Worker {
	return &Worker{
		store:    st,
		queue:    q,
		probe:    probe.NewClient(),
		analyzer: analyzer.New(),
		enc:      enc,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	slog.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		payload, err := w.queue.Dequeue(ctx, 5*time.Second)
		if err != nil {
			slog.Error("dequeue failed", "error", err)
			continue
		}
		if payload == nil {
			continue
		}

		if err := w.processJob(ctx, *payload); err != nil {
			slog.Error("process job failed", "job_id", payload.JobID, "error", err)
		}
	}
}

func (w *Worker) processJob(ctx context.Context, payload queue.JobPayload) error {
	relay, err := w.store.GetRelay(ctx, payload.RelayID)
	if err != nil {
		return err
	}
	if relay == nil {
		return fmt.Errorf("relay not found: %s", payload.RelayID)
	}

	if err := w.store.UpdateJobStatus(ctx, payload.JobID, store.JobRunning, nil); err != nil {
		return err
	}

	var procErr error
	switch payload.Type {
	case store.JobHealth:
		procErr = w.runHealthProbe(ctx, relay, payload)
	case store.JobPerformance:
		procErr = w.runPerformanceProbe(ctx, relay, payload)
	case store.JobAuthenticity:
		procErr = w.runAuthenticityProbe(ctx, relay, payload)
	default:
		procErr = fmt.Errorf("unknown job type: %s", payload.Type)
	}

	if procErr != nil {
		msg := procErr.Error()
		_ = w.store.UpdateJobStatus(ctx, payload.JobID, store.JobFailed, &msg)
		return procErr
	}
	return w.store.UpdateJobStatus(ctx, payload.JobID, store.JobCompleted, nil)
}

func (w *Worker) endpoint(relay *store.Relay) (probe.Endpoint, error) {
	if len(relay.APIKeyEncrypted) == 0 {
		return probe.Endpoint{}, fmt.Errorf("relay %s has no API key configured", relay.Name)
	}
	key, err := w.enc.Decrypt(relay.APIKeyEncrypted)
	if err != nil {
		return probe.Endpoint{}, err
	}
	return probe.Endpoint{
		BaseURL: relay.APIBaseURL,
		APIKey:  key,
		Backups: relay.BackupAPIBaseURLs,
	}, nil
}

func (w *Worker) runHealthProbe(ctx context.Context, relay *store.Relay, payload queue.JobPayload) error {
	ep, err := w.endpoint(relay)
	if err != nil {
		return w.saveErrorResult(ctx, payload, relay, relay.HealthModel, store.JobHealth, err)
	}

	model := relay.HealthModel
	prompt := "Reply with OK"

	streamRes, err := w.probe.ChatCompletion(ctx, ep, model, prompt, true)
	if err != nil {
		return err
	}
	if err := w.saveProbeResult(ctx, payload, relay, model, store.JobHealth, true, streamRes); err != nil {
		return err
	}

	nonStreamRes, err := w.probe.ChatCompletion(ctx, ep, model, prompt, false)
	if err != nil {
		return err
	}
	return w.saveProbeResult(ctx, payload, relay, model, store.JobHealth, false, nonStreamRes)
}

func (w *Worker) runPerformanceProbe(ctx context.Context, relay *store.Relay, payload queue.JobPayload) error {
	ep, err := w.endpoint(relay)
	if err != nil {
		return w.saveErrorResult(ctx, payload, relay, payload.Model, store.JobPerformance, err)
	}

	model := payload.Model
	if model == "" && len(relay.ClaimedModels) > 0 {
		model = relay.ClaimedModels[0]
	}
	prompt := "Explain what an API relay is in two short sentences."

	streamRes, err := w.probe.ChatCompletion(ctx, ep, model, prompt, true)
	if err != nil {
		return err
	}
	if err := w.saveProbeResult(ctx, payload, relay, model, store.JobPerformance, true, streamRes); err != nil {
		return err
	}

	nonStreamRes, err := w.probe.ChatCompletion(ctx, ep, model, prompt, false)
	if err != nil {
		return err
	}
	if err := w.saveProbeResult(ctx, payload, relay, model, store.JobPerformance, false, nonStreamRes); err != nil {
		return err
	}

	return w.store.UpsertDailyAggregate(ctx, relay.ID, model, time.Now())
}

func (w *Worker) runAuthenticityProbe(ctx context.Context, relay *store.Relay, payload queue.JobPayload) error {
	ep, err := w.endpoint(relay)
	if err != nil {
		return err
	}

	model := payload.Model
	if model == "" && len(relay.ClaimedModels) > 0 {
		model = relay.ClaimedModels[0]
	}

	cases, err := w.store.ListPromptCases(ctx)
	if err != nil {
		return err
	}

	var promptResults []analyzer.PromptResult
	var lastResponseModel string

	for _, c := range cases {
		res, err := w.probe.ChatCompletion(ctx, ep, model, c.Prompt, false)
		if err != nil {
			continue
		}
		promptResults = append(promptResults, analyzer.PromptResult{Case: c, Response: res})
		if res.ResponseModel != "" {
			lastResponseModel = res.ResponseModel
		}
		if err := w.saveProbeResult(ctx, payload, relay, model, store.JobAuthenticity, false, res); err != nil {
			return err
		}
	}

	report := w.analyzer.Analyze(analyzer.AnalysisInput{
		ClaimedModel:  model,
		ResponseModel: lastResponseModel,
		PromptResults: promptResults,
	})
	report.RelayID = relay.ID
	jobID := payload.JobID
	report.JobID = &jobID

	return w.store.InsertAuthenticityReport(ctx, &report)
}

func (w *Worker) saveErrorResult(ctx context.Context, payload queue.JobPayload, relay *store.Relay, model string, probeType store.JobType, err error) error {
	msg := err.Error()
	res := &probe.Result{Error: msg}
	return w.saveProbeResult(ctx, payload, relay, model, probeType, false, res)
}

func (w *Worker) saveProbeResult(ctx context.Context, payload queue.JobPayload, relay *store.Relay, model string, probeType store.JobType, stream bool, res *probe.Result) error {
	var errPtr *string
	if res.Error != "" {
		errPtr = &res.Error
	}
	var respModel *string
	if res.ResponseModel != "" {
		respModel = &res.ResponseModel
	}
	var hash *string
	if res.ResponseHash != "" {
		hash = &res.ResponseHash
	}
	summary := probe.Summarize(res.Content, 200)
	var summaryPtr *string
	if summary != "" {
		summaryPtr = &summary
	}
	status := res.HTTPStatus
	if status == 0 && res.Error != "" {
		status = 0
	}

	pr := &store.ProbeResult{
		JobID:           payload.JobID,
		RelayID:         relay.ID,
		Model:           model,
		ProbeType:       probeType,
		Stream:          stream,
		LatencyMS:       intPtr(res.LatencyMS),
		TTFTMS:          res.TTFTMS,
		TPOTMS:          res.TPOTMS,
		InputTokens:     res.InputTokens,
		OutputTokens:    res.OutputTokens,
		HTTPStatus:      intPtr(status),
		ResponseModel:   respModel,
		Error:           errPtr,
		RawResponseHash: hash,
		ResponseSummary: summaryPtr,
	}
	return w.store.InsertProbeResult(ctx, pr)
}

func intPtr(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}
