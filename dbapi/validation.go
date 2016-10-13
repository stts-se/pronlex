package dbapi

// For validating a lexicon db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stts-se/pronlex/lex"
	"github.com/stts-se/pronlex/validation"
)

type ValStats struct {
	InvalidEntries   int
	TotalEntries     int
	TotalValidations int
	Levels           map[string]int `json:levels`
	Rules            map[string]int `json:rules`
}

func processChunk(db *sql.DB, chunk []int64, vd validation.Validator, stats ValStats) (ValStats, error) {
	q := Query{EntryIDs: chunk}
	var w lex.EntrySliceWriter

	err := LookUp(db, q, &w)
	if err != nil {
		return stats, fmt.Errorf("couldn't lookup from ids : %s", err)
	}
	updated := []lex.Entry{}
	for i, e := range w.Entries {
		oldVal := e.EntryValidations
		e, _ = vd.ValidateEntry(e)
		newVal := e.EntryValidations
		if len(newVal) > 0 {
			stats.InvalidEntries += 1
			for _, v := range newVal {
				stats.TotalValidations += 1
				stats.Levels[strings.ToLower(v.Level)] += 1
				stats.Rules[strings.ToLower(v.RuleName+" ("+v.Level+")")] += 1
			}
		}
		w.Entries[i] = e
		if len(oldVal) > 0 || len(newVal) > 0 {
			updated = append(updated, e)
		}
	}
	err = UpdateValidation(db, updated)
	if err != nil {
		return stats, fmt.Errorf("couldn't update validation : %s", err)
	}
	return stats, nil
}

func Validate(db *sql.DB, logger Logger, vd validation.Validator, q Query) (ValStats, error) {

	stats := ValStats{Levels: make(map[string]int), Rules: make(map[string]int)}

	logger.Write(fmt.Sprintf("query: %v", q))

	q.PageLength = 0 //todo?
	q.Page = 0       //todo?

	logger.Write("Fetching entries from lexicon ... ")
	ids, err := LookUpIds(db, q)
	if err != nil {
		return stats, fmt.Errorf("couldn't lookup for validation : %s", err)
	}
	total := len(ids)
	stats.TotalEntries = total
	stats.InvalidEntries = 0
	stats.TotalValidations = 0
	logger.Write(fmt.Sprintf("Found %d entries", total))

	n := 0

	chunkSize := 500
	var chunk []int64

	for _, id := range ids {
		n = n + 1
		chunk = append(chunk, id)

		if n%chunkSize == 0 {
			stats, err = processChunk(db, chunk, vd, stats)
			if err != nil {
				return stats, err
			}
			chunk = []int64{}

		}

		if n%10 == 0 {
			js, err := json.Marshal(stats)
			if err != nil {
				return stats, fmt.Errorf("couldn't marshal validation stats : %s", err)
			}
			msg := fmt.Sprintf("%s", js)
			logger.Write(msg)
		}
	}
	if len(chunk) > 0 {
		stats, err = processChunk(db, chunk, vd, stats)
		if err != nil {
			return stats, err
		}
		chunk = []int64{}
	}

	return stats, nil
}
