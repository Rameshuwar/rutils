package nifty

import (
	"context"
	"log"
	"time"
)

const (
	// TargetSyncHourIST is 10 (10:00 AM)
	TargetSyncHourIST = 10
	// TargetSyncMinuteIST is 0
	TargetSyncMinuteIST = 0
)

// Scheduler manages in-process scheduled background synchronization of NIFTY 50 data
type Scheduler struct {
	service  *Service
	location *time.Location
	stopChan chan struct{}
}

// NewScheduler creates a new Scheduler
func NewScheduler(service *Service) *Scheduler {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}
	return &Scheduler{
		service:  service,
		location: loc,
		stopChan: make(chan struct{}),
	}
}

// NeedsSync determines if a sync is needed right now (e.g. on startup or after wake from sleep)
func NeedsSync(now time.Time, lastData *NiftyData, loc *time.Location) bool {
	if lastData == nil || len(lastData.Companies) == 0 {
		return true
	}

	nowIST := now.In(loc)
	today10AM := time.Date(nowIST.Year(), nowIST.Month(), nowIST.Day(), TargetSyncHourIST, TargetSyncMinuteIST, 0, 0, loc)

	// If it's currently before 10:00 AM IST today:
	// If lastData was updated before yesterday's 10:00 AM IST (or more than 24 hours ago), we can also consider it old,
	// but if it has any valid data, we wait until 10:00 AM IST today.
	if nowIST.Before(today10AM) {
		return false
	}

	// It is now at or after 10:00 AM IST today.
	// Parse the last updated timestamp:
	lastUpdated, err := time.Parse(time.RFC3339, lastData.UpdatedAt)
	if err != nil {
		return true
	}
	lastUpdatedIST := lastUpdated.In(loc)

	// If the last update occurred before today's 10:00 AM IST, then today's 10:00 AM sync was missed!
	if lastUpdatedIST.Before(today10AM) {
		return true
	}

	return false
}

// NextScheduledRun calculates the next 10:00 AM IST time
func NextScheduledRun(now time.Time, loc *time.Location) time.Time {
	nowIST := now.In(loc)
	today10AM := time.Date(nowIST.Year(), nowIST.Month(), nowIST.Day(), TargetSyncHourIST, TargetSyncMinuteIST, 0, 0, loc)

	if nowIST.Before(today10AM) {
		return today10AM
	}

	// Next run is tomorrow at 10:00 AM IST
	tomorrow := nowIST.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), TargetSyncHourIST, TargetSyncMinuteIST, 0, 0, loc)
}

// Start runs the scheduler in the background. It returns immediately and operates asynchronously.
func (s *Scheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop terminates the background scheduler
func (s *Scheduler) Stop() {
	select {
	case <-s.stopChan:
	default:
		close(s.stopChan)
	}
}

func (s *Scheduler) run(ctx context.Context) {
	log.Printf("[NIFTY 50] In-process scheduler started (Timezone: Asia/Kolkata, Schedule: Daily at 10:00 AM IST)")

	// 1. Startup Check: Did we miss a sync because the system was shut down at 10:00 AM?
	now := time.Now().In(s.location)
	lastData, err := s.service.GetCompanies()
	if err != nil || NeedsSync(now, lastData, s.location) {
		if lastData == nil {
			log.Printf("[NIFTY 50] No existing dataset found on startup. Executing initial sync...")
		} else {
			log.Printf("[NIFTY 50] System startup detected missed 10:00 AM IST sync (Last updated: %s). Executing catch-up sync...", lastData.UpdatedAt)
		}

		syncCtx, syncCancel := context.WithTimeout(ctx, 45*time.Second)
		if _, syncErr := s.service.Sync(syncCtx); syncErr != nil {
			log.Printf("[NIFTY 50] Startup sync error: %v", syncErr)
		}
		syncCancel()
	} else {
		log.Printf("[NIFTY 50] Today's dataset is up-to-date (Last updated: %s). Next sync at %s",
			lastData.UpdatedAt, NextScheduledRun(now, s.location).Format("2006-01-02 15:04 MST"))
	}

	// 2. Watchdog loop with timer + 1-minute tick check to ensure robust execution even if clock shifts
	watchdog := time.NewTicker(1 * time.Minute)
	defer watchdog.Stop()

	for {
		now = time.Now().In(s.location)
		nextRun := NextScheduledRun(now, s.location)
		waitDuration := time.Until(nextRun)
		if waitDuration < 0 {
			waitDuration = 0
		}

		select {
		case <-ctx.Done():
			log.Printf("[NIFTY 50] Scheduler stopped (context canceled)")
			return
		case <-s.stopChan:
			log.Printf("[NIFTY 50] Scheduler stopped")
			return
		case <-time.After(waitDuration):
			log.Printf("[NIFTY 50] Scheduled 10:00 AM IST trigger activated. Starting daily sync...")
			syncCtx, syncCancel := context.WithTimeout(ctx, 45*time.Second)
			if _, syncErr := s.service.Sync(syncCtx); syncErr != nil {
				log.Printf("[NIFTY 50] Scheduled sync error: %v", syncErr)
			}
			syncCancel()

		case <-watchdog.C:
			// Fallback check in case system was suspended or time changed
			nowCheck := time.Now().In(s.location)
			currentData, _ := s.service.GetCompanies()
			if NeedsSync(nowCheck, currentData, s.location) {
				log.Printf("[NIFTY 50] Watchdog detected pending 10:00 AM IST sync. Executing sync...")
				syncCtx, syncCancel := context.WithTimeout(ctx, 45*time.Second)
				if _, syncErr := s.service.Sync(syncCtx); syncErr != nil {
					log.Printf("[NIFTY 50] Watchdog sync error: %v", syncErr)
				}
				syncCancel()
			}
		}
	}
}
