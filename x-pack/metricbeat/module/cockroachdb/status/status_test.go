// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// skipping tests on windows 32 bit versions, not supported
//go:build !integration && !windows && !386

package status

import (
	"os"
	"testing"

	"github.com/elastic/beats/v7/metricbeat/mb"
	mbtest "github.com/elastic/beats/v7/metricbeat/mb/testing"
	"github.com/elastic/elastic-agent-libs/logp"

	// Register input module and metricset
	_ "github.com/elastic/beats/v7/metricbeat/module/prometheus"
	_ "github.com/elastic/beats/v7/metricbeat/module/prometheus/collector"
)

func init() {
	// To be moved to some kind of helper
	os.Setenv("BEAT_STRICT_PERMS", "false")
	mb.Registry.SetSecondarySource(mb.NewLightModulesSource(logp.NewNopLogger(), "../../../module"))
}

func TestEventMapping(t *testing.T) {

	mbtest.TestDataFiles(t, "cockroachdb", "status")
}
