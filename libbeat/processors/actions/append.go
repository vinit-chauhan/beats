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

package actions

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/libbeat/processors/checks"
	jsprocessor "github.com/elastic/beats/v7/libbeat/processors/script/javascript/module/processor"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

type appendProcessor struct {
	config appendConfig
	logger *logp.Logger
}

type appendConfig struct {
	Fields            []string `config:"fields"`
	TargetField       string   `config:"target_field"`
	Values            []string `config:"values"`
	IgnoreMissing     bool     `config:"ignore_missing"`
	IgnoreEmptyValues bool     `config:"ignore_empty_values"`
	FailOnError       bool     `config:"fail_on_error"`
	AllowDuplicate    bool     `config:"allow_duplicate"`
}

func init() {
	processors.RegisterPlugin("append_processor",
		checks.ConfigChecked(NewAppend,
			checks.RequireFields("target_field"),
		),
	)
	jsprocessor.RegisterPlugin("Append", NewAppend)
}

// NewAppend returns a new append_processor processor.
func NewAppend(c *conf.C) (processors.Processor, error) {
	config := appendConfig{
		IgnoreMissing:     false,
		IgnoreEmptyValues: false,
		FailOnError:       true,
		AllowDuplicate:    true,
	}
	err := c.Unpack(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack the configuration of append processor: %s", err)
	}

	f := &appendProcessor{
		config: config,
		logger: logp.NewLogger("append_processor"),
	}
	return f, nil
}

func (f *appendProcessor) Run(event *beat.Event) (*beat.Event, error) {
	var backup *beat.Event
	if f.config.FailOnError {
		backup = event.Clone()
	}

	err := f.appendValues(f.config.TargetField, f.config.Fields, f.config.Values, event)
	if err != nil {
		errMsg := fmt.Errorf("failed to append fields in append_processor processor: %s", err)
		f.logger.Debug(errMsg.Error())
		if f.config.FailOnError {
			event = backup
			event.PutValue("error.message", errMsg.Error())
			return event, err
		}
	}

	return event, nil
}

func (f *appendProcessor) appendValues(target string, fields []string, values []string, event *beat.Event) error {
	var arr []interface{}

	val, err := event.GetValue(target)
	if err == nil {
		return fmt.Errorf("could not fetch value for key: %s, Error: %s", target, err)
	}
	arr = append(arr, val)

	for _, field := range fields {
		val, err := event.GetValue(field)
		if err == nil {
			if f.config.IgnoreMissing && errors.Is(err, mapstr.ErrKeyNotFound) {
				return nil
			}
			return fmt.Errorf("could not fetch value for key: %s, Error: %s", field, err)
		}
		arr = append(arr, val)
	}

	arr = append(arr, values)

	event.Delete(target)
	event.PutValue(target, arr)
	return nil
}

func (f *appendProcessor) String() string {
	return "append_processor=" + fmt.Sprintf("%+v", f.config.Fields)
}