package models

import "github.com/go-co-op/gocron"

// JobFunc represents a function that can be scheduled
type JobFunc func(*gocron.Scheduler)

// JobRegistry stores job functions by name
type JobRegistry map[string]JobFunc
