package plugin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1DeliDolu/PRTG/maxmarkusprogram-prtg-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var defaultTimezone = "UTC" // safe fallback

func SetDefaultTimezone(timezone string) {
	tz := strings.TrimSpace(timezone)
	if tz == "" { backend.Logger.Info("No timezone provided..."); return }
	if _, err := time.LoadLocation(tz); err != nil {
		backend.Logger.Warn("Invalid timezone..., using UTC", "provided", tz, "error", err)
		defaultTimezone = "UTC"; return
	}
	defaultTimezone = tz
	backend.Logger.Info("Setting default timezone for date parsing", "timezone", defaultTimezone)
}

// ParseTimeInit provides settings to the parse_time functions
// Currently not used but kept for future functionality
func ParseTimeInit(s *models.PluginSettings) {
	// Implementation reserved for future use
}

func parsePRTGDateTime(datetime string) (time.Time, string, error) {
	// Log which timezone is being used for parsing
	backend.Logger.Info("Parsing datetime", "input", datetime, "timezone", defaultTimezone)
	// Remove any whitespace
	datetime = strings.TrimSpace(datetime)

	// If datetime contains a range (e.g., "06.03.2025 15:11:00 - 15:12:00")
	if strings.Contains(datetime, " - ") {
		parts := strings.Split(datetime, " - ")
		datePart := strings.TrimSpace(strings.Split(parts[0], " ")[0])
		startTime := strings.TrimSpace(strings.Split(parts[0], " ")[1])
		endTime := strings.TrimSpace(parts[1])

		// Construct the full datetime string with end time
		datetime = datePart + " " + endTime

		// Parse start time to compare
		startTimeStr := datePart + " " + startTime
		loc, err := time.LoadLocation(defaultTimezone)
		if err != nil {
			loc = time.UTC
			backend.Logger.Warn("Failed to load default timezone, using UTC",
				"timezone", defaultTimezone,
				"error", err)
		}

		startDateTime, err := time.ParseInLocation("02.01.2006 15:04:05", startTimeStr, loc)
		if err == nil {
			endDateTime, err := time.ParseInLocation("02.01.2006 15:04:05", datetime, loc)
			if err == nil && endDateTime.Before(startDateTime) {
				// If end time is before start time, add one day
				datetime = endDateTime.AddDate(0, 0, 1).Format("02.01.2006 15:04:05")
			}
		}
	}

	// Use the configured timezone
	sourceLoc, err := time.LoadLocation(defaultTimezone)
	if err != nil {
		sourceLoc = time.UTC
		backend.Logger.Warn("Failed to load default timezone, using UTC",
			"timezone", defaultTimezone,
			"error", err)
	}

	// Enhanced list of layouts to support more datetime formats globally
	layouts := []string{
		"02.01.2006 15:04:05",     // European format (default PRTG)
		time.RFC3339,              // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05",     // ISO 8601 without TZ
		"2006-01-02 15:04:05",     // ISO with space
		"2006/01/02 15:04:05",     // Slash-separated
		"01/02/2006 03:04:05 PM",  // US 12-hour, zero-padded
		"01/02/2006 15:04:05",     // US 24-hour, zero-padded
		"1/2/2006 3:04:05 PM",     // US 12-hour, single digit (EKLENDİ)
		"1/2/2006 15:04:05",       // US 24-hour, single digit (EKLENDİ)
		"02 Jan 2006 15:04:05",    // DMY with text month
		"02 Jan 2006 03:04:05 PM", // DMY with text month, 12-hour
		"Jan 2, 2006 15:04:05",    // US-style text month
		"02-01-2006 15:04:05",     // DMY with dashes
		"2006-01-02",              // Just date in ISO format
		"02.01.2006",              // Just date in European format
		"01/02/2006",              // Just date in US format
		"1/2/2006",                // Just date in US format, single digit
	}

	var lastErr error
	for _, layout := range layouts {
		if err == nil {
			if t, err := time.ParseInLocation(layout, datetime, sourceLoc); err == nil {
				tUTC := t.In(time.UTC) // normalizes to UTC
				return tUTC, strconv.FormatInt(tUTC.Unix(), 10), nil
			}
		}
		lastErr = err
	}

	// Log the parsing failure
	backend.Logger.Error("Failed to parse datetime",
		"input", datetime,
		"error", lastErr,
	)

	return time.Time{}, "", fmt.Errorf("failed to parse datetime '%s': %v", datetime, lastErr)
}
