package kite

import "time"

const (
	defaultPublicStaticDir = "static"
	shutDownTimeout        = 30 * time.Second
	kiteTraceExporter      = "kite"
	kiteTracerURL          = "https://tracer.github.com/sllt/kite"
	checkPortTimeout       = 2 * time.Second
	kiteHost               = "https://github.com/sllt/kite"
	startServerPing        = "/api/ping/up"
	shutServerPing         = "/api/ping/down"
	pingTimeout            = 5 * time.Second
	defaultTelemetry       = "true"
	defaultReflection      = "false"
)
