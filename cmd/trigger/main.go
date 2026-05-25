package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wadeling/origin-check/internal/config"
	"github.com/wadeling/origin-check/internal/queue"
	"github.com/wadeling/origin-check/internal/store"
	"github.com/wadeling/origin-check/internal/trigger"
)

func main() {
	var (
		listRelays = flag.Bool("list", false, "list active relays and exit")
		relayName  = flag.String("relay", "", "relay name (required unless -list)")
		jobType    = flag.String("type", "authenticity", "job type: authenticity, performance, health")
		model      = flag.String("model", "", "single model only (default: all claimed models)")
		wait       = flag.Bool("wait", true, "wait for worker to finish jobs")
		timeout    = flag.Duration("timeout", 15*time.Minute, "max wait when -wait")
	)
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fatal(err)
	}
	defer st.Close()

	if *listRelays {
		relays, err := st.ListActiveRelays(ctx)
		if err != nil {
			fatal(err)
		}
		for _, r := range relays {
			fmt.Printf("%s\tmodels=%s\n", r.Name, strings.Join(r.ClaimedModels, ","))
		}
		return
	}

	if *relayName == "" {
		fmt.Fprintln(os.Stderr, "usage: trigger -relay <name> [-type authenticity] [-model gpt-5.5] [-wait] [-list]")
		os.Exit(2)
	}

	q, err := queue.New(cfg.RedisURL)
	if err != nil {
		fatal(err)
	}
	defer q.Close()

	relay, err := st.GetRelayByName(ctx, *relayName)
	if err != nil {
		fatal(err)
	}
	if relay == nil {
		fatal(fmt.Errorf("relay not found: %q (use -list)", *relayName))
	}

	jt, err := parseJobType(*jobType)
	if err != nil {
		fatal(err)
	}

	models := trigger.ModelsForJob(*relay, jt, strings.TrimSpace(*model))
	jobs, err := trigger.Enqueue(ctx, st, q, trigger.Options{
		Relay:   *relay,
		JobType: jt,
		Models:  models,
	})
	if err != nil {
		fatal(err)
	}

	for _, j := range jobs {
		fmt.Printf("enqueued\trelay=%s\ttype=%s\tmodel=%s\tjob_id=%s\n", relay.Name, jt, j.Model, j.JobID)
	}

	if !*wait {
		fmt.Println("jobs queued; worker will process them asynchronously")
		return
	}

	fmt.Printf("waiting for %d job(s) (timeout %s)...\n", len(jobs), *timeout)
	ids := make([]uuid.UUID, len(jobs))
	for i, j := range jobs {
		ids[i] = j.JobID
	}

	finished, err := trigger.Wait(ctx, st, ids, *timeout)
	if err != nil {
		fatal(err)
	}

	failed := 0
	for _, job := range finished {
		modelLabel := ""
		if job.Model != nil {
			modelLabel = *job.Model
		}
		switch job.Status {
		case store.JobCompleted:
			fmt.Printf("completed\tmodel=%s\tjob_id=%s\n", modelLabel, job.ID)
		case store.JobFailed:
			failed++
			errMsg := ""
			if job.Error != nil {
				errMsg = *job.Error
			}
			fmt.Printf("failed\tmodel=%s\tjob_id=%s\terror=%s\n", modelLabel, job.ID, errMsg)
		}
	}

	if jt == store.JobAuthenticity {
		printAuthenticityReports(ctx, st, relay.ID, models)
	}

	if failed > 0 {
		os.Exit(1)
	}
}

func parseJobType(s string) (store.JobType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "authenticity", "auth":
		return store.JobAuthenticity, nil
	case "performance", "perf":
		return store.JobPerformance, nil
	case "health":
		return store.JobHealth, nil
	default:
		return "", fmt.Errorf("unknown job type %q (use authenticity, performance, health)", s)
	}
}

func printAuthenticityReports(ctx context.Context, st *store.Store, relayID uuid.UUID, models []string) {
	reports, err := st.ListAuthenticityReports(ctx, relayID, models, len(models))
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: list reports: %v\n", err)
		return
	}
	for _, r := range reports {
		fmt.Printf("report\tmodel=%s\tscore=%.1f\tverdict=%s\tat=%s\n",
			r.ClaimedModel, r.Score, r.Verdict, r.CreatedAt.Format(time.RFC3339))
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
