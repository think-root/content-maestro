package models

import "github.com/go-co-op/gocron"

type JobFunc func(*gocron.Scheduler)

type JobRegistry map[string]JobFunc
