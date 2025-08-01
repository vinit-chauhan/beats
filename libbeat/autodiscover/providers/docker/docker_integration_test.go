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

//go:build (linux || darwin || windows) && integration

package docker

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/beats/v7/libbeat/autodiscover/template"
	dk "github.com/elastic/beats/v7/libbeat/tests/docker"
	"github.com/elastic/elastic-agent-autodiscover/bus"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/keystore"
	"github.com/elastic/elastic-agent-libs/logp/logptest"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

// Test docker start emits an autodiscover event
func TestDockerStart(t *testing.T) {
	log := logptest.NewTestingLogger(t, "docker")

	d, err := dk.NewClient(log)
	if err != nil {
		t.Fatal(err)
	}

	UUID, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}
	bus := bus.New(log, "test")
	config := defaultConfig()
	config.CleanupTimeout = 0

	s := &template.MapperSettings{nil, nil}
	config.Templates = *s
	k, _ := keystore.NewFileKeystore("test")
	provider, err := AutodiscoverBuilder("mockBeat", bus, UUID, conf.MustNewConfigFrom(config), k, log)
	if err != nil {
		t.Fatal(err)
	}

	provider.Start()
	defer provider.Stop()

	listener := bus.Subscribe()
	defer listener.Stop()

	// Start
	cmd := []string{"echo", "Hi!"}
	labels := map[string]string{"label": "foo", "label.child": "bar"}
	ID, err := d.ContainerStart("busybox:latest", cmd, labels)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := d.ContainerRemove(ID); err != nil {
			t.Log(err)
		}
	}()

	checkEvent(t, listener, ID, true)

	// Kill
	if err := d.ContainerKill(ID); err != nil {
		t.Log(err)
	}
	checkEvent(t, listener, ID, false)
}

func getValue(e bus.Event, key string) interface{} {
	val, err := mapstr.M(e).GetValue(key)
	if err != nil {
		return nil
	}
	return val
}

func checkEvent(t *testing.T, listener bus.Listener, id string, start bool) {
	timeout := time.After(60 * time.Second)
	for {
		select {
		case e := <-listener.Events():
			// Ignore any other container
			if getValue(e, "container.id") != id {
				continue
			}
			if start {
				assert.Equal(t, getValue(e, "start"), true)
				assert.Nil(t, getValue(e, "stop"))
			} else {
				assert.Equal(t, getValue(e, "stop"), true)
				assert.Nil(t, getValue(e, "start"))
			}
			assert.Equal(t, getValue(e, "container.image.name"), "busybox:latest")
			// labels.dedot=true by default
			assert.Equal(t,
				mapstr.M{
					"label": mapstr.M{
						"value": "foo",
						"child": "bar",
					},
				},
				getValue(e, "container.labels"),
			)
			assert.NotNil(t, getValue(e, "container.id"))
			assert.NotNil(t, getValue(e, "container.name"))
			assert.NotNil(t, getValue(e, "host"))
			assert.Equal(t, getValue(e, "docker.container.id"), getValue(e, "meta.container.id"))
			assert.Equal(t, getValue(e, "docker.container.name"), getValue(e, "meta.container.name"))
			assert.Equal(t, getValue(e, "docker.container.image"), getValue(e, "meta.container.image.name"))
			return
		case <-timeout:
			t.Fatal("Timeout waiting for provider events")
			return
		}
	}
}
