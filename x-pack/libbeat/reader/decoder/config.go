// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package decoder

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// decoderConfig contains the configuration options for instantiating a decoder.
type Config struct {
	Codec *CodecConfig `config:"codec"`
}

// codecConfig contains the configuration options for different codecs used by a decoder.
type CodecConfig struct {
	CSV     *CSVCodecConfig     `config:"csv"`
	Parquet *ParquetCodecConfig `config:"parquet"`
}

func (c *CodecConfig) Validate() error {
	count := 0
	if c == nil {
		return nil
	}

	if c.Parquet != nil {
		count++
	}
	if c.CSV != nil {
		count++
	}

	if count > 1 {
		return errors.New("more than one decoder configured")
	}

	return nil
}

// CSVCodecConfig contains the configuration options for the CSV codec.
type CSVCodecConfig struct {
	Enabled bool `config:"enabled"`

	// Fields is the set of field names. If it is present
	// it is used to specify the object names of returned
	// values and the FieldsPerRecord field in the csv.Reader.
	// Otherwise, names are obtained from the first
	// line of the CSV data.
	Fields []string `config:"fields_names"`

	// The fields below have the same meaning as the
	// fields of the same name in csv.Reader.
	Comma            *Rune `config:"comma"`
	Comment          Rune  `config:"comment"`
	LazyQuotes       bool  `config:"lazy_quotes"`
	TrimLeadingSpace bool  `config:"trim_leading_space"`
}

type Rune rune

func (r *Rune) Unpack(s string) error {
	if s == "" {
		return nil
	}
	n := utf8.RuneCountInString(s)
	if n != 1 {
		return fmt.Errorf("single character option given more than one character: %q", s)
	}
	_r, _ := utf8.DecodeRuneInString(s)
	*r = Rune(_r)
	return nil
}

// parquetCodecConfig contains the configuration options for the parquet codec.
type ParquetCodecConfig struct {
	Enabled bool `config:"enabled"`

	// If ProcessParallel is true, then functions which read multiple columns will read those columns in parallel
	// from the file with a number of readers equal to the number of columns. Otherwise columns are read serially.
	ProcessParallel bool `config:"process_parallel"`
	// BatchSize is the number of rows to read at a time from the file.
	BatchSize int `config:"batch_size" default:"1"`
}
