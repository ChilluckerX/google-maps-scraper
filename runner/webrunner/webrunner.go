package webrunner

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosom/google-maps-scraper/deduper"
	"github.com/gosom/google-maps-scraper/exiter"
	"github.com/gosom/google-maps-scraper/runner"
	"github.com/gosom/google-maps-scraper/tlmt"
	"github.com/gosom/google-maps-scraper/web"
	"github.com/gosom/google-maps-scraper/web/sqlite"
	"github.com/gosom/scrapemate"
	"github.com/gosom/scrapemate/adapters/writers/csvwriter"
	"github.com/gosom/scrapemate/scrapemateapp"
	"golang.org/x/sync/errgroup"
)

type webrunner struct {
	srv *web.Server
	svc *web.Service
	cfg *runner.Config
}

func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg.DataFolder == "" {
		return nil, fmt.Errorf("data folder is required")
	}

	if err := os.MkdirAll(cfg.DataFolder, os.ModePerm); err != nil {
		return nil, err
	}

	const dbfname = "jobs.db"

	dbpath := filepath.Join(cfg.DataFolder, dbfname)

	repo, err := sqlite.New(dbpath)
	if err != nil {
		return nil, err
	}

	svc := web.NewService(repo, cfg.DataFolder)

	srv, err := web.New(svc, cfg.Addr)
	if err != nil {
		return nil, err
	}

	ans := webrunner{
		srv: srv,
		svc: svc,
		cfg: cfg,
	}

	return &ans, nil
}

func (w *webrunner) Run(ctx context.Context) error {
	egroup, ctx := errgroup.WithContext(ctx)

	egroup.Go(func() error {
		return w.work(ctx)
	})

	egroup.Go(func() error {
		return w.srv.Start(ctx)
	})

	return egroup.Wait()
}

func (w *webrunner) Close(context.Context) error {
	return nil
}

func (w *webrunner) work(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			jobs, err := w.svc.SelectPending(ctx)
			if err != nil {
				return err
			}

			for i := range jobs {
				select {
				case <-ctx.Done():
					return nil
				default:
					t0 := time.Now().UTC()
					if err := w.scrapeJob(ctx, &jobs[i]); err != nil {
						params := map[string]any{
							"job_count": len(jobs[i].Data.Keywords),
							"duration":  time.Now().UTC().Sub(t0).String(),
							"error":     err.Error(),
						}

						evt := tlmt.NewEvent("web_runner", params)

						_ = runner.Telemetry().Send(ctx, evt)

						log.Printf("error scraping job %s: %v", jobs[i].ID, err)
					} else {
						params := map[string]any{
							"job_count": len(jobs[i].Data.Keywords),
							"duration":  time.Now().UTC().Sub(t0).String(),
						}

						_ = runner.Telemetry().Send(ctx, tlmt.NewEvent("web_runner", params))

						log.Printf("job %s scraped successfully", jobs[i].ID)
					}
				}
			}
		}
	}
}

func (w *webrunner) scrapeJob(ctx context.Context, job *web.Job) error {
	job.Status = web.StatusWorking

	err := w.svc.Update(ctx, job)
	if err != nil {
		return err
	}

	if len(job.Data.Keywords) == 0 {
		job.Status = web.StatusFailed

		return w.svc.Update(ctx, job)
	}

	var isRescrape bool
	var oldPath string

	if job.Data.MergeWithJobId != "" {
		isRescrape = true
		oldPath = filepath.Join(w.cfg.DataFolder, job.Data.MergeWithJobId+".csv")
	}

	outpath := filepath.Join(w.cfg.DataFolder, job.ID+".csv")
	var scrapeOutPath string = outpath

	if isRescrape {
		scrapeOutPath = filepath.Join(w.cfg.DataFolder, job.ID+"_temp.csv")
	}

	outfile, err := os.Create(scrapeOutPath)
	if err != nil {
		return err
	}

	defer func() {
		_ = outfile.Close()
	}()

	mate, err := w.setupMate(ctx, outfile, job)
	if err != nil {
		job.Status = web.StatusFailed

		err2 := w.svc.Update(ctx, job)
		if err2 != nil {
			log.Printf("failed to update job status: %v", err2)
		}

		return err
	}

	defer mate.Close()

	var coords string
	if job.Data.Lat != "" && job.Data.Lon != "" {
		coords = job.Data.Lat + "," + job.Data.Lon
	}

	dedup := deduper.New()
	exitMonitor := exiter.New()

	seedJobs, err := runner.CreateSeedJobs(
		job.Data.FastMode,
		job.Data.Lang,
		strings.NewReader(strings.Join(job.Data.Keywords, "\n")),
		job.Data.Depth,
		job.Data.Email,
		coords,
		job.Data.Zoom,
		func() float64 {
			if job.Data.Radius <= 0 {
				return 10000 // 10 km
			}

			return float64(job.Data.Radius)
		}(),
		dedup,
		exitMonitor,
		w.cfg.ExtraReviews,
	)
	if err != nil {
		err2 := w.svc.Update(ctx, job)
		if err2 != nil {
			log.Printf("failed to update job status: %v", err2)
		}

		return err
	}

	if len(seedJobs) > 0 {
		exitMonitor.SetSeedCount(len(seedJobs))

		allowedSeconds := max(60, len(seedJobs)*10*job.Data.Depth/50+120)

		if job.Data.MaxTime > 0 {
			if job.Data.MaxTime.Seconds() < 180 {
				allowedSeconds = 180
			} else {
				allowedSeconds = int(job.Data.MaxTime.Seconds())
			}
		}

		log.Printf("running job %s with %d seed jobs and %d allowed seconds", job.ID, len(seedJobs), allowedSeconds)

		mateCtx, cancel := context.WithTimeout(ctx, time.Duration(allowedSeconds)*time.Second)
		defer cancel()

		exitMonitor.SetCancelFunc(cancel)

		go exitMonitor.Run(mateCtx)

		err = mate.Start(mateCtx, seedJobs...)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			cancel()

			err2 := w.svc.Update(ctx, job)
			if err2 != nil {
				log.Printf("failed to update job status: %v", err2)
			}

			return err
		}

		cancel()
	}

	mate.Close()
	outfile.Close()

	if isRescrape {
		if err := w.mergeCSV(oldPath, scrapeOutPath, outpath); err != nil {
			log.Printf("Failed to merge CSV: %v", err)
		} else {
			os.Remove(scrapeOutPath)
		}
	}

	job.Status = web.StatusOK

	return w.svc.Update(ctx, job)
}

func (w *webrunner) setupMate(_ context.Context, writer io.Writer, job *web.Job) (*scrapemateapp.ScrapemateApp, error) {
	opts := []func(*scrapemateapp.Config) error{
		scrapemateapp.WithConcurrency(w.cfg.Concurrency),
		scrapemateapp.WithExitOnInactivity(time.Minute * 3),
	}

	if !job.Data.FastMode {
		opts = append(opts,
			scrapemateapp.WithJS(scrapemateapp.DisableImages()),
		)
	} else {
		opts = append(opts,
			scrapemateapp.WithStealth("firefox"),
		)
	}

	hasProxy := false

	if len(w.cfg.Proxies) > 0 {
		opts = append(opts, scrapemateapp.WithProxies(w.cfg.Proxies))
		hasProxy = true
	} else if len(job.Data.Proxies) > 0 {
		opts = append(opts,
			scrapemateapp.WithProxies(job.Data.Proxies),
		)
		hasProxy = true
	}

	if !w.cfg.DisablePageReuse {
		opts = append(opts,
			scrapemateapp.WithPageReuseLimit(2),
			scrapemateapp.WithPageReuseLimit(200),
		)
	}

	log.Printf("job %s has proxy: %v", job.ID, hasProxy)

	csvWriter := csvwriter.NewCsvWriter(csv.NewWriter(writer))

	writers := []scrapemate.ResultWriter{csvWriter}

	matecfg, err := scrapemateapp.NewConfig(
		writers,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return scrapemateapp.NewScrapeMateApp(matecfg)
}

func (w *webrunner) mergeCSV(oldPath, newPath, outPath string) error {
	// Read old CSV
	oldFile, err := os.Open(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If old file doesn't exist, just copy the new one over
			return os.Rename(newPath, outPath)
		}
		return err
	}
	oldReader := csv.NewReader(oldFile)
	oldRecords, err := oldReader.ReadAll()
	oldFile.Close()
	if err != nil {
		return err
	}

	// Read new CSV
	newFile, err := os.Open(newPath)
	if err != nil {
		return err
	}
	newReader := csv.NewReader(newFile)
	newRecords, err := newReader.ReadAll()
	newFile.Close()
	if err != nil {
		return err
	}

	if len(oldRecords) == 0 {
		oldRecords = newRecords
	} else if len(newRecords) > 1 {
		// Use "link" as key to merge
		linkIdx := -1
		for i, h := range oldRecords[0] {
			if h == "link" {
				linkIdx = i
				break
			}
		}

		if linkIdx == -1 {
			linkIdx = 1 // Default to column 1 if header check fails
		}

		merged := make(map[string][]string)
		for i := 1; i < len(oldRecords); i++ {
			if len(oldRecords[i]) > linkIdx {
				merged[oldRecords[i][linkIdx]] = oldRecords[i]
			}
		}

		for i := 1; i < len(newRecords); i++ {
			if len(newRecords[i]) > linkIdx {
				merged[newRecords[i][linkIdx]] = newRecords[i]
			}
		}

		finalRecords := [][]string{oldRecords[0]}
		for _, rec := range merged {
			finalRecords = append(finalRecords, rec)
		}
		oldRecords = finalRecords
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	return writer.WriteAll(oldRecords)
}
