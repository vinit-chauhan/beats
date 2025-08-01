// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package add_nomad_metadata

import (
	"fmt"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common/cfgwarn"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/nomad"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

type nomadAnnotator struct {
	watcher  nomad.Watcher
	indexers *Indexers
	matchers *Matchers
	cache    *cache
}

func init() {
	processors.RegisterPlugin("add_nomad_metadata", New)

	// Register default indexers
	Indexing.AddIndexer(AllocationNameIndexerName, NewAllocationNameIndexer)
	Indexing.AddIndexer(AllocationUUIDIndexerName, NewAllocationUUIDIndexer)
	Indexing.AddMatcher(FieldMatcherName, NewFieldMatcher)
	Indexing.AddMatcher(FieldFormatMatcherName, NewFieldFormatMatcher)
}

// New constructs a new add_nomad_metadata processor.
func New(cfg *conf.C, log *logp.Logger) (beat.Processor, error) {
	log.Warn(cfgwarn.Experimental("The add_nomad_metadata processor is experimental"))

	config := defaultNomadAnnotatorConfig()

	err := cfg.Unpack(&config)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the nomad configuration: %w", err)
	}

	//Load default indexer configs
	if config.DefaultIndexers.Enabled {
		Indexing.RLock()
		for key, cfg := range Indexing.GetDefaultIndexerConfigs() {
			config.Indexers = append(config.Indexers, map[string]conf.C{key: cfg})
		}
		Indexing.RUnlock()
	}

	//Load default matcher configs
	if config.DefaultMatchers.Enabled {
		Indexing.RLock()
		for key, cfg := range Indexing.GetDefaultMatcherConfigs() {
			config.Matchers = append(config.Matchers, map[string]conf.C{key: cfg})
		}
		Indexing.RUnlock()
	}

	clientConfig := nomad.ClientConfig{
		Address:   config.Address,
		Namespace: config.Namespace,
		Region:    config.Region,
		SecretID:  config.SecretID,
	}
	client, err := nomad.NewClient(clientConfig)
	if err != nil {
		log.Error("nomad: Couldn't create client")
		return nil, err
	}

	metaGen, err := nomad.NewMetaGenerator(cfg, client)
	if err != nil {
		return nil, err
	}

	indexers := NewIndexers(config.Indexers, metaGen)
	matchers := NewMatchers(config.Matchers, log)

	log.Named("nomad").Debugf("Using node: %s", config.Node)
	log.Named("nomad").Debug("Initializing watcher")

	options := nomad.WatchOptions{
		SyncTimeout:     config.syncPeriod,
		RefreshInterval: config.RefreshInterval,
		Namespace:       config.Namespace,
	}
	if config.Scope == ScopeNode {
		node := config.Node
		if node == "" {
			agent, err := client.Agent().Self()
			if err != nil {
				return nil, fmt.Errorf("`scope: %s` used without `node`: couldn't autoconfigure node name: %w", ScopeNode, err)
			}
			if agent.Member.Name == "" {
				return nil, fmt.Errorf("`scope: %s` used without `node`: API returned empty name", ScopeNode)
			}
			node = agent.Member.Name
		}
		options.Node = node
	}
	watcher, err := nomad.NewWatcher(client, options, log)
	if err != nil {
		log.Errorf("Error creating watcher %v", err.Error())
		return nil, err
	}

	processor := &nomadAnnotator{
		watcher:  watcher,
		indexers: indexers,
		matchers: matchers,
		cache:    newCache(config.CleanupTimeout),
	}

	watcher.AddEventHandler(nomad.ResourceEventHandlerFuncs{
		AddFunc: func(alloc nomad.Resource) {
			processor.addAllocation(&alloc)
		},
		DeleteFunc: func(alloc nomad.Resource) {
			processor.removeAllocation(&alloc)
		},
		UpdateFunc: func(alloc nomad.Resource) {
			processor.removeAllocation(&alloc)
			processor.addAllocation(&alloc)
		},
	})

	if err := watcher.Start(); err != nil {
		return nil, err
	}

	return processor, nil
}

func (n *nomadAnnotator) Run(event *beat.Event) (*beat.Event, error) {
	index := n.matchers.MetadataIndex(event.Fields)
	if index == "" {
		return event, nil
	}

	metadata := n.cache.get(index)
	if metadata == nil {
		return event, nil
	}

	event.Fields.DeepUpdate(mapstr.M{
		"nomad": metadata.Clone(),
	})

	return event, nil
}

func (n *nomadAnnotator) addAllocation(alloc *nomad.Resource) {
	metadata := n.indexers.GetMetadata(alloc)

	for _, m := range metadata {
		n.cache.set(m.Index, m.Data)
	}
}

func (n *nomadAnnotator) removeAllocation(alloc *nomad.Resource) {
	metadata := n.indexers.GetIndexes(alloc)

	for _, idx := range metadata {
		n.cache.delete(idx)
	}
}

func (n *nomadAnnotator) String() string {
	return "add_nomad_metadata"
}
