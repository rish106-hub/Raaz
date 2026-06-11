package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	wsConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "raaz_ws_connections_active",
		Help: "Current number of active WebSocket connections.",
	})

	matchQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "raaz_match_queue_depth",
		Help: "Current number of clients waiting for a match.",
	})

	matchesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "raaz_matches_total",
		Help: "Total sessions created since process start.",
	})

	moderationStrikesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "raaz_moderation_strikes_total",
		Help: "Moderation strikes recorded, labelled by content category.",
	}, []string{"category"})

	crisisTriggersTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "raaz_crisis_triggers_total",
		Help: "Crisis support triggers since process start.",
	})
)
