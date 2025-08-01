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

package input

import (
	"fmt"
	"sync"
	"time"

	"github.com/elastic/beats/v7/filebeat/channel"
	"github.com/elastic/beats/v7/filebeat/input/file"
	"github.com/elastic/beats/v7/libbeat/management/status"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/monitoring"
)

var inputList = monitoring.NewUniqueList()
var inputListMetricsOnce sync.Once

// RegisterMonitoringInputs registers a list of inputs with the
// monitoring system under the provided namespace.  If namespace is
// empty, it default to "state". Registration only occurs once.
func RegisterMonitoringInputs(namespace string) {
	if namespace == "" {
		namespace = "state"
	}
	inputListMetricsOnce.Do(func() {
		monitoring.NewFunc(monitoring.GetNamespace(namespace).GetRegistry(), "input", inputList.Report, monitoring.Report)
	})
}

// Input is the interface common to all input
type Input interface {
	Run()
	Stop()
	Wait()
}

// Runner encapsulate the lifecycle of the input
type Runner struct {
	config         inputConfig
	input          Input
	done           chan struct{}
	wg             *sync.WaitGroup
	Once           bool
	beatDone       chan struct{}
	statusReporter status.StatusReporter
	logger         *logp.Logger
}

// New instantiates a new Runner
func New(
	conf *conf.C,
	connector channel.Connector,
	beatDone chan struct{},
	states []file.State,
	logger *logp.Logger,
) (*Runner, error) {
	input := &Runner{
		config:   defaultConfig,
		wg:       &sync.WaitGroup{},
		done:     make(chan struct{}),
		Once:     false,
		beatDone: beatDone,
		logger:   logger,
	}

	var err error
	if err = conf.Unpack(&input.config); err != nil {
		return nil, err
	}

	// log config related warnings
	input.config.checkUnsupportedConfig(logger)

	var f Factory
	f, err = GetFactory(input.config.Type)
	if err != nil {
		return input, err
	}

	context := Context{
		States:            states,
		Done:              input.done,
		BeatDone:          input.beatDone,
		Meta:              nil,
		GetStatusReporter: input.GetStatusReporter,
	}
	var ipt Input
	ipt, err = f(conf, connector, context, logger)
	if err != nil {
		return input, err
	}
	input.input = ipt

	return input, nil
}

// Start starts the input
func (p *Runner) Start() {
	p.wg.Add(1)

	onceWg := sync.WaitGroup{}
	if p.Once {
		// Make sure start is only completed when Run did a complete first scan
		defer onceWg.Wait()
	}

	onceWg.Add(1)
	inputList.Add(p.config.Type)
	// Add waitgroup to make sure input is finished
	go func() {
		defer func() {
			onceWg.Done()
			p.stop()
			p.wg.Done()
		}()

		p.Run()
	}()
}

// Run starts scanning through all the file paths and fetch the related files. Start a harvester for each file
func (p *Runner) Run() {
	// Initial input run
	p.input.Run()

	// Shuts down after the first complete run of all input
	if p.Once {
		return
	}

	for {
		select {
		case <-p.done:
			p.logger.Info("input ticker stopped")
			return
		case <-time.After(p.config.ScanFrequency):
			p.logger.Named("input").Debug("Run input")
			p.input.Run()
		}
	}
}

// Stop stops the input and with it all harvesters
func (p *Runner) Stop() {
	// Stop scanning and wait for completion
	close(p.done)
	p.wg.Wait()
	inputList.Remove(p.config.Type)
}

func (p *Runner) stop() {
	// In case of once, it will be waited until harvesters close itself
	if p.Once {
		p.input.Wait()
	} else {
		p.input.Stop()
	}
}

func (p *Runner) String() string {
	return fmt.Sprintf("input [type=%s]", p.config.Type)
}

func (p *Runner) SetStatusReporter(statusReporter status.StatusReporter) {
	p.statusReporter = statusReporter
}

func (p *Runner) GetStatusReporter() status.StatusReporter {
	return p.statusReporter
}
