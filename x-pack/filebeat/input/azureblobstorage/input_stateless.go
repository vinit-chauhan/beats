// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package azureblobstorage

import (
	"context"

	"golang.org/x/sync/errgroup"

	v2 "github.com/elastic/beats/v7/filebeat/input/v2"
	cursor "github.com/elastic/beats/v7/filebeat/input/v2/input-cursor"
	stateless "github.com/elastic/beats/v7/filebeat/input/v2/input-stateless"
	"github.com/elastic/beats/v7/libbeat/beat"
)

type statelessInput struct {
	config     config
	serviceURL string
}

func (statelessInput) Name() string {
	return "azure-blob-storage-stateless"
}

func newStatelessInput(config config, url string) *statelessInput {
	return &statelessInput{config: config, serviceURL: url}
}

func (in *statelessInput) Test(v2.TestContext) error {
	return nil
}

type statelessPublisher struct {
	wrapped stateless.Publisher
}

func (pub statelessPublisher) Publish(event beat.Event, _ interface{}) error {
	pub.wrapped.Publish(event)
	return nil
}

// Run starts the input and blocks until it ends the execution.
func (in *statelessInput) Run(inputCtx v2.Context, publisher stateless.Publisher) error {
	pub := statelessPublisher{wrapped: publisher}
	var source cursor.Source
	var g errgroup.Group
	for _, c := range in.config.Containers {
		container := tryOverrideOrDefault(in.config, c)
		source = &Source{
			AccountName:              in.config.AccountName,
			ContainerName:            c.Name,
			BatchSize:                *container.BatchSize,
			MaxWorkers:               *container.MaxWorkers,
			Poll:                     *container.Poll,
			PollInterval:             *container.PollInterval,
			TimeStampEpoch:           container.TimeStampEpoch,
			ExpandEventListFromField: container.ExpandEventListFromField,
			FileSelectors:            container.FileSelectors,
			ReaderConfig:             container.ReaderConfig,
			PathPrefix:               container.PathPrefix,
		}

		st := newState()
		currentSource := source.(*Source)
		log := inputCtx.Logger.With("account_name", currentSource.AccountName).With("container", currentSource.ContainerName)
		// create a new inputMetrics instance
		metrics := newInputMetrics(inputCtx.ID+":"+currentSource.ContainerName, nil)
		metrics.url.Set(in.serviceURL + currentSource.ContainerName)
		defer metrics.Close()

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-inputCtx.Cancelation.Done()
			cancel()
		}()

		serviceClient, credential, err := fetchServiceClientAndCreds(in.config, in.serviceURL, log)
		if err != nil {
			return err
		}
		containerClient, err := fetchContainerClient(serviceClient, currentSource.ContainerName, log)
		if err != nil {
			return err
		}

		scheduler := newScheduler(pub, containerClient, credential, currentSource, &in.config, st, in.serviceURL, noopReporter{}, metrics, log)
		// allows multiple containers to be scheduled concurrently while testing
		// the stateless input is triggered only while testing and till now it did not mimic
		// the real world concurrent execution of multiple containers. This fix allows it to do so.
		g.Go(func() error {
			return scheduler.schedule(ctx)
		})

	}
	return g.Wait()
}
