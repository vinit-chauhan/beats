// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build linux || darwin || windows

package network

import (
	"fmt"

	"github.com/docker/docker/client"

	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/elastic/beats/v7/metricbeat/module/docker"
)

func init() {
	mb.Registry.MustAddMetricSet("docker", "network", New,
		mb.WithHostParser(docker.HostParser),
		mb.DefaultMetricSet(),
	)
}

type MetricSet struct {
	mb.BaseMetricSet
	netService   *NetService
	dockerClient *client.Client
	dedot        bool
}

// New creates a new instance of the docker network MetricSet.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	config := docker.DefaultConfig()
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	client, err := docker.NewDockerClient(base.HostData().URI, config, base.Logger())
	if err != nil {
		return nil, err
	}

	return &MetricSet{
		BaseMetricSet: base,
		dockerClient:  client,
		netService: &NetService{
			NetworkStatPerContainer: make(map[string]map[string]NetRaw),
		},
		dedot: config.DeDot,
	}, nil
}

// Fetch methods creates a list of network events for each container.
func (m *MetricSet) Fetch(r mb.ReporterV2) error {
	stats, err := docker.FetchStats(m.dockerClient, m.Module().Config().Timeout, false, m.Logger())
	if err != nil {
		return fmt.Errorf("failed to get docker stats: %w", err)
	}

	formattedStats := m.netService.getNetworkStatsPerContainer(stats, m.dedot)
	eventsMapping(r, formattedStats)

	return nil
}

// Close stops the metricset
func (m *MetricSet) Close() error {

	return m.dockerClient.Close()
}
