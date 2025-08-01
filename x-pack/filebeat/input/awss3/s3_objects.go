// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package awss3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/reader"
	"github.com/elastic/beats/v7/libbeat/reader/readfile"
	"github.com/elastic/beats/v7/libbeat/reader/readfile/encoding"
	x_reader "github.com/elastic/beats/v7/x-pack/libbeat/reader"
	"github.com/elastic/beats/v7/x-pack/libbeat/reader/decoder"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

type s3ObjectProcessorFactory struct {
	metrics       *inputMetrics
	s3            s3API
	fileSelectors []fileSelectorConfig
	backupConfig  backupConfig
}

type s3ObjectProcessor struct {
	*s3ObjectProcessorFactory

	ctx           context.Context
	eventCallback func(beat.Event)
	readerConfig  *readerConfig // Config about how to process the object.
	s3Obj         s3EventV2     // S3 object information.
	s3ObjHash     string
	s3RequestURL  string

	s3Metadata map[string]interface{} // S3 object metadata.
}

type s3DownloadedObject struct {
	body        io.ReadCloser
	contentType string
	metadata    map[string]interface{}
}

const (
	contentTypeJSON   = "application/json"
	contentTypeNDJSON = "application/x-ndjson"
)

// errS3DownloadFailed reports problems downloading an S3 object. Download errors
// should never treated as permanent, they are just an indication to apply a
// retry backoff until the connection is healthy again.
var errS3DownloadFailed = errors.New("S3 download failure")

func newS3ObjectProcessorFactory(metrics *inputMetrics, s3 s3API, sel []fileSelectorConfig, backupConfig backupConfig) *s3ObjectProcessorFactory {
	if metrics == nil {
		// Metrics are optional. Initialize a stub.
		metrics = newInputMetrics("", nil, 0)
	}
	if len(sel) == 0 {
		sel = []fileSelectorConfig{
			{ReaderConfig: defaultConfig().ReaderConfig},
		}
	}
	return &s3ObjectProcessorFactory{
		metrics:       metrics,
		s3:            s3,
		fileSelectors: sel,
		backupConfig:  backupConfig,
	}
}

func (f *s3ObjectProcessorFactory) findReaderConfig(key string) *readerConfig {
	for _, sel := range f.fileSelectors {
		if sel.Regex == nil || sel.Regex.MatchString(key) {
			return &sel.ReaderConfig
		}
	}
	return nil
}

// Create returns a new s3ObjectProcessor. It returns nil when no file selectors
// match the S3 object key.
func (f *s3ObjectProcessorFactory) Create(ctx context.Context, obj s3EventV2) s3ObjectHandler {
	readerConfig := f.findReaderConfig(obj.S3.Object.Key)
	if readerConfig == nil {
		// No file_selectors are a match, skip.
		return nil
	}

	return &s3ObjectProcessor{
		s3ObjectProcessorFactory: f,
		ctx:                      ctx,
		readerConfig:             readerConfig,
		s3Obj:                    obj,
		s3ObjHash:                s3ObjectHash(obj),
	}
}

func (p *s3ObjectProcessor) ProcessS3Object(log *logp.Logger, eventCallback func(e beat.Event)) error {
	if p == nil {
		return nil
	}
	p.eventCallback = eventCallback
	log = log.With(
		"bucket_arn", p.s3Obj.S3.Bucket.Name,
		"object_key", p.s3Obj.S3.Object.Key,
		"last_modified", p.s3Obj.S3.Object.LastModified)

	// Metrics and Logging
	log.Debug("Begin S3 object processing.")
	p.metrics.s3ObjectsRequestedTotal.Inc()
	p.metrics.s3ObjectsInflight.Inc()
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		p.metrics.s3ObjectsInflight.Dec()
		p.metrics.s3ObjectProcessingTime.Update(elapsed.Nanoseconds())
		log.Debugw("End S3 object processing.", "elapsed_time_ns", elapsed)
	}()

	// Request object (download).
	s3Obj, err := p.download()
	if err != nil {
		// Wrap downloadError in the result so the caller knows it's not a
		// permanent failure.
		return fmt.Errorf("%w: %w", errS3DownloadFailed, err)
	}
	defer s3Obj.body.Close()

	p.s3Metadata = s3Obj.metadata

	mReader := newMonitoredReader(s3Obj.body, p.metrics.s3BytesProcessedTotal)

	// Wrap to detect S3 body streaming errors so we can retry them
	wrappedReader := s3DownloadFailedWrappedReader{r: mReader}

	streamReader, err := x_reader.AddGzipDecoderIfNeeded(wrappedReader)
	if err != nil {
		return fmt.Errorf("failed checking for gzip content: %w", err)
	}

	// Overwrite with user configured Content-Type.
	if p.readerConfig.ContentType != "" {
		s3Obj.contentType = p.readerConfig.ContentType
	}

	// try to create a dec from the using the codec config
	dec, err := decoder.NewDecoder(p.readerConfig.Decoding, streamReader)
	if err != nil {
		return err
	}
	switch dec := dec.(type) {
	case decoder.ValueDecoder:
		defer dec.Close()

		for dec.Next() {
			evtOffset, msg, _, err := dec.DecodeValue()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				break
			}
			evt := p.createEvent(string(msg), evtOffset)

			p.eventCallback(evt)
		}

	case decoder.Decoder:
		var evtOffset int64
		defer dec.Close()

		for dec.Next() {
			data, err := dec.Decode()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				break
			}
			evtOffset, err = p.readJSONSlice(bytes.NewReader(data), evtOffset)
			if err != nil {
				break
			}
		}

	default:
		// This is the legacy path. It will be removed in future and clubbed together with the decoder.
		// Process object content stream.
		switch {
		case strings.HasPrefix(s3Obj.contentType, contentTypeJSON) || strings.HasPrefix(s3Obj.contentType, contentTypeNDJSON):
			err = p.readJSON(streamReader)
		default:
			err = p.readFile(streamReader, log)
		}
	}
	if err != nil {
		return fmt.Errorf("failed reading s3 object (elapsed_time_ns=%d): %w",
			time.Since(start).Nanoseconds(), err)
	}

	// finally obtain total bytes of the object through metered reader
	p.metrics.s3ObjectSizeInBytes.Update(mReader.totalBytesReadCurrent)

	return nil
}

// download requests the S3 object from AWS and returns the object's
// Content-Type and reader to get the object's contents.
// The caller must close the reader embedded in s3DownloadedObject.
func (p *s3ObjectProcessor) download() (obj *s3DownloadedObject, err error) {
	getObjectOutput, err := p.s3.GetObject(p.ctx, p.s3Obj.AWSRegion, p.s3Obj.S3.Bucket.Name, p.s3Obj.S3.Object.Key)
	if err != nil {
		return nil, err
	}
	if getObjectOutput == nil {
		return nil, fmt.Errorf("empty response from s3 get object: %w", err)
	}
	s3RequestURL := getObjectOutput.ResultMetadata.Get(s3RequestURLMetadataKey)
	if s3RequestURLAsString, ok := s3RequestURL.(string); ok {
		p.s3RequestURL = s3RequestURLAsString
	}

	meta := s3Metadata(getObjectOutput, p.readerConfig.IncludeS3Metadata...)

	ctType := ""
	if getObjectOutput.ContentType != nil {
		ctType = *getObjectOutput.ContentType
	}

	s := &s3DownloadedObject{
		body:        getObjectOutput.Body,
		contentType: ctType,
		metadata:    meta,
	}

	return s, nil
}

func (p *s3ObjectProcessor) readJSON(r io.Reader) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	for dec.More() && p.ctx.Err() == nil {
		offset := dec.InputOffset()

		var item json.RawMessage
		if err := dec.Decode(&item); err != nil {
			return fmt.Errorf("failed to decode json: %w", err)
		}

		if p.readerConfig.ExpandEventListFromField != "" {
			if err := p.splitEventList(p.readerConfig.ExpandEventListFromField, item, offset, p.s3ObjHash); err != nil {
				return err
			}
			continue
		}

		data, _ := item.MarshalJSON()
		evt := p.createEvent(string(data), offset)
		p.eventCallback(evt)
	}

	return nil
}

// readJSONSlice uses a json.RawMessage to process JSON slice data as individual JSON objects.
// It accepts a reader and a starting offset, it returns an updated offset and an error if any.
// It reads the opening token separately and then iterates over the slice, decoding each object and publishing it.
func (p *s3ObjectProcessor) readJSONSlice(r io.Reader, evtOffset int64) (int64, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	// reads starting token separately since this is always a slice.
	_, err := dec.Token()
	if err != nil {
		return -1, fmt.Errorf("failed to read JSON slice token for object key: %s, with error: %w", p.s3Obj.S3.Object.Key, err)
	}

	// we track each event offset separately since we are reading a slice.
	for dec.More() && p.ctx.Err() == nil {
		var item json.RawMessage
		if err := dec.Decode(&item); err != nil {
			return -1, fmt.Errorf("failed to decode json: %w", err)
		}

		if p.readerConfig.ExpandEventListFromField != "" {
			if err := p.splitEventList(p.readerConfig.ExpandEventListFromField, item, evtOffset, p.s3ObjHash); err != nil {
				return -1, err
			}
			continue
		}

		data, _ := item.MarshalJSON()
		evt := p.createEvent(string(data), evtOffset)
		p.eventCallback(evt)
		evtOffset++
	}

	return evtOffset, nil
}

func (p *s3ObjectProcessor) splitEventList(key string, raw json.RawMessage, offset int64, objHash string) error {
	// .[] signifies the root object is an array, and it should be split.
	if key != ".[]" {
		var jsonObject map[string]json.RawMessage
		if err := json.Unmarshal(raw, &jsonObject); err != nil {
			return err
		}

		var found bool
		raw, found = jsonObject[key]
		if !found {
			return fmt.Errorf("expand_event_list_from_field key <%v> is not in event", key)
		}
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return fmt.Errorf("expand_event_list_from_field <%v> is not an array", key)
	}

	for dec.More() {
		arrayOffset := dec.InputOffset()

		var item json.RawMessage
		if err := dec.Decode(&item); err != nil {
			return fmt.Errorf("failed to decode array item at offset %d: %w", offset+arrayOffset, err)
		}

		data, _ := item.MarshalJSON()
		p.s3ObjHash = objHash
		evt := p.createEvent(string(data), offset+arrayOffset)
		p.eventCallback(evt)
	}

	return nil
}

func (p *s3ObjectProcessor) readFile(r io.Reader, logger *logp.Logger) error {
	encodingFactory, ok := encoding.FindEncoding(p.readerConfig.Encoding)
	if !ok || encodingFactory == nil {
		return fmt.Errorf("failed to find '%v' encoding", p.readerConfig.Encoding)
	}

	enc, err := encodingFactory(r)
	if err != nil {
		return fmt.Errorf("failed to initialize encoding: %w", err)
	}

	var reader reader.Reader
	reader, err = readfile.NewEncodeReader(io.NopCloser(r), readfile.Config{
		Codec:        enc,
		BufferSize:   int(p.readerConfig.BufferSize),
		Terminator:   p.readerConfig.LineTerminator,
		CollectOnEOF: true,
		MaxBytes:     int(p.readerConfig.MaxBytes) * 4,
	}, logger)
	if err != nil {
		return fmt.Errorf("failed to create encode reader: %w", err)
	}

	reader = readfile.NewStripNewline(reader, p.readerConfig.LineTerminator)
	reader = p.readerConfig.Parsers.Create(reader, logger)
	reader = readfile.NewLimitReader(reader, int(p.readerConfig.MaxBytes))

	var offset int64
	for {
		message, err := reader.Next()
		if len(message.Content) > 0 {
			event := p.createEvent(string(message.Content), offset)
			event.Fields.DeepUpdate(message.Fields)
			offset += int64(message.Bytes)
			p.eventCallback(event)
		}

		if errors.Is(err, io.EOF) {
			// No more lines
			break
		}
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}
	}

	return nil
}

// createEvent constructs a beat.Event from message and offset. The value of
// message populates the event message field, and offset is used to set the
// log.offset field and, with the object's ARN and key, the @metadata._id field.
// If offset is negative, it is ignored. No @metadata._id field is added to
// the event and the log.offset field is not set.
func (p *s3ObjectProcessor) createEvent(message string, offset int64) beat.Event {
	event := beat.Event{
		Timestamp: time.Now().UTC(),
		Fields: mapstr.M{
			"message": message,
			"log": mapstr.M{
				"file": mapstr.M{
					"path": p.s3RequestURL,
				},
			},
			"aws": mapstr.M{
				"s3": mapstr.M{
					"bucket": mapstr.M{
						"name": p.s3Obj.S3.Bucket.Name,
						"arn":  p.s3Obj.S3.Bucket.ARN,
					},
					"object": mapstr.M{
						"key": p.s3Obj.S3.Object.Key,
					},
				},
			},
			"cloud": mapstr.M{
				"provider": p.s3Obj.Provider,
				"region":   p.s3Obj.AWSRegion,
			},
		},
	}
	if offset >= 0 {
		event.Fields.Put("log.offset", offset)
		event.SetID(objectID(p.s3Obj.S3.Object.LastModified, p.s3ObjHash, offset))
	}

	if len(p.s3Metadata) > 0 {
		_, _ = event.Fields.Put("aws.s3.metadata", p.s3Metadata)
	}

	return event
}

func (p *s3ObjectProcessor) FinalizeS3Object() error {
	bucketName := p.backupConfig.GetBucketName()
	if bucketName == "" {
		return nil
	}
	backupKey := p.backupConfig.BackupToBucketPrefix + p.s3Obj.S3.Object.Key
	_, err := p.s3.CopyObject(p.ctx, p.s3Obj.AWSRegion, p.s3Obj.S3.Bucket.Name, bucketName, p.s3Obj.S3.Object.Key, backupKey)
	if err != nil {
		return fmt.Errorf("failed to copy object to backup bucket: %w", err)
	}
	if !p.backupConfig.Delete {
		return nil
	}
	_, err = p.s3.DeleteObject(p.ctx, p.s3Obj.AWSRegion, p.s3Obj.S3.Bucket.Name, p.s3Obj.S3.Object.Key)
	if err != nil {
		return fmt.Errorf("failed to delete object from bucket: %w", err)
	}
	return nil
}

func objectID(lastModified time.Time, objectHash string, offset int64) string {
	return fmt.Sprintf("%d-%s-%012d", lastModified.UnixNano(), objectHash, offset)
}

// s3ObjectHash returns a short sha256 hash of the bucket arn + object key name.
func s3ObjectHash(obj s3EventV2) string {
	h := sha256.New()
	h.Write([]byte(obj.S3.Bucket.ARN))
	h.Write([]byte(obj.S3.Object.Key))
	prefix := hex.EncodeToString(h.Sum(nil))
	return prefix[:10]
}

// s3Metadata returns a map containing the selected S3 object metadata keys.
func s3Metadata(resp *s3.GetObjectOutput, keys ...string) mapstr.M {
	if len(keys) == 0 {
		return nil
	}

	// When you upload objects using the REST API, the optional user-defined
	// metadata names must begin with "x-amz-meta-" to distinguish them from
	// other HTTP headers.
	const userMetaPrefix = "x-amz-meta-"

	allMeta := map[string]interface{}{}

	// Get headers using AWS SDK struct tags.
	fields := reflect.TypeOf(resp).Elem()
	values := reflect.ValueOf(resp).Elem()
	for i := 0; i < fields.NumField(); i++ {
		f := fields.Field(i)

		if loc, _ := f.Tag.Lookup("location"); loc != "header" {
			continue
		}

		name, found := f.Tag.Lookup("locationName")
		if !found {
			continue
		}
		name = strings.ToLower(name)

		if name == userMetaPrefix {
			continue
		}

		v := values.Field(i)
		switch v.Kind() {
		case reflect.Ptr:
			if v.IsNil() {
				continue
			}
			v = v.Elem()
		default:
			if v.IsZero() {
				continue
			}
		}

		allMeta[name] = v.Interface()
	}

	// Add in the user defined headers.
	for k, v := range resp.Metadata {
		k = strings.ToLower(k)
		allMeta[userMetaPrefix+k] = v
	}

	// Select the matching headers from the config.
	metadata := mapstr.M{}
	for _, key := range keys {
		key = strings.ToLower(key)

		v, found := allMeta[key]
		if !found {
			continue
		}

		metadata[key] = v
	}

	return metadata
}

// s3DownloadFailedWrappedReader implements io.Reader interface.
// Internally, it validates Read errors for io.ErrUnexpectedEOF and wraps them with errS3DownloadFailed.
type s3DownloadFailedWrappedReader struct {
	r io.Reader
}

func (s3r s3DownloadFailedWrappedReader) Read(p []byte) (n int, err error) {
	n, err = s3r.r.Read(p)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return n, fmt.Errorf("%w: %w", errS3DownloadFailed, err)
	}

	return n, err
}
