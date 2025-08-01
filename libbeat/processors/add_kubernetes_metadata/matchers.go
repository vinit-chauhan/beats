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

package add_kubernetes_metadata

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common/fmtstr"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/outputs/codec/format"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

const (
	FieldMatcherName       = "fields"
	FieldFormatMatcherName = "field_format"
	regexKeyGroupName      = "key"
)

// Matcher takes a new event and returns the index
type Matcher interface {
	// MetadataIndex returns the index string to use in annotation lookups for the given
	// event. A previous indexer should have generated that index for this to work
	// This function can return "" if the event doesn't match
	MetadataIndex(event mapstr.M) string
}

type Matchers struct {
	matchers []Matcher
}

type MatcherConstructor func(config config.C, logger *logp.Logger) (Matcher, error)

func NewMatchers(configs PluginConfig, logger *logp.Logger) *Matchers {
	matchers := []Matcher{}
	for _, pluginConfigs := range configs {
		for name, pluginConfig := range pluginConfigs {
			matchFunc := Indexing.GetMatcher(name)
			if matchFunc == nil {
				logger.Warnf("Unable to find matcher plugin %s", name)
				continue
			}

			matcher, err := matchFunc(pluginConfig, logger)
			if err != nil {
				logger.Warnf("Unable to initialize matcher plugin %s due to error %v", name, err)
				continue
			}

			matchers = append(matchers, matcher)

		}
	}
	return &Matchers{
		matchers: matchers,
	}
}

// MetadataIndex returns the index string for the first matcher from the Registry returning one
func (m *Matchers) MetadataIndex(event mapstr.M) string {
	for _, matcher := range m.matchers {
		index := matcher.MetadataIndex(event)
		if index != "" {
			return index
		}
	}

	// No index returned
	return ""
}

func (m *Matchers) Empty() bool {
	return len(m.matchers) == 0
}

type FieldMatcher struct {
	MatchFields []string
	Regexp      *regexp.Regexp
}

func NewFieldMatcher(cfg config.C, _ *logp.Logger) (Matcher, error) {
	matcherConfig := struct {
		LookupFields []string `config:"lookup_fields"`
		RegexPattern string   `config:"regex_pattern"`
	}{}

	err := cfg.Unpack(&matcherConfig)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the fields matcher configuration: %w", err)
	}

	if len(matcherConfig.LookupFields) == 0 {
		return nil, fmt.Errorf("lookup_fields can not be empty")
	}

	if len(matcherConfig.RegexPattern) == 0 {
		return &FieldMatcher{MatchFields: matcherConfig.LookupFields}, nil
	}
	regex, err := regexp.Compile(matcherConfig.RegexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	captureGroupNames := regex.SubexpNames()
	if !slices.Contains(captureGroupNames, regexKeyGroupName) {
		return nil, fmt.Errorf("regex missing required capture group `key`")
	}

	return &FieldMatcher{MatchFields: matcherConfig.LookupFields, Regexp: regex}, nil
}

func (f *FieldMatcher) MetadataIndex(event mapstr.M) string {
	for _, field := range f.MatchFields {
		fieldIface, err := event.GetValue(field)
		if err != nil {
			continue
		}
		fieldValue, ok := fieldIface.(string)
		if !ok {
			continue
		}
		if f.Regexp == nil {
			return fieldValue
		}

		matches := f.Regexp.FindStringSubmatch(fieldValue)
		if matches == nil {
			continue
		}
		keyIndex := f.Regexp.SubexpIndex(regexKeyGroupName)
		key := matches[keyIndex]
		if key != "" {
			return key
		}
	}

	return ""
}

type FieldFormatMatcher struct {
	Codec codec.Codec
}

func NewFieldFormatMatcher(cfg config.C, _ *logp.Logger) (Matcher, error) {
	config := struct {
		Format string `config:"format"`
	}{}

	err := cfg.Unpack(&config)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the `format` configuration of `field_format` matcher: %w", err)
	}

	if config.Format == "" {
		return nil, fmt.Errorf("`format` of `field_format` matcher can't be empty")
	}

	return &FieldFormatMatcher{
		Codec: format.New(fmtstr.MustCompileEvent(config.Format)),
	}, nil

}

func (f *FieldFormatMatcher) MetadataIndex(event mapstr.M) string {
	bytes, err := f.Codec.Encode("", &beat.Event{
		Fields: event,
	})

	if err != nil {
		logp.Debug("kubernetes", "Unable to apply field format pattern on event")
	}

	if len(bytes) == 0 {
		return ""
	}

	return string(bytes)
}
