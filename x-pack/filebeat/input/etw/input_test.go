// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

//go:build windows

package etw

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"

	input "github.com/elastic/beats/v7/filebeat/input/v2"
	"github.com/elastic/beats/v7/x-pack/libbeat/reader/etw"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

type mockSessionOperator struct {
	// Fields to store function implementations that tests can customize
	newSessionFunc              func(config config) (*etw.Session, error)
	attachToExistingSessionFunc func(session *etw.Session) error
	createRealtimeSessionFunc   func(session *etw.Session) error
	startConsumerFunc           func(session *etw.Session) error
	stopSessionFunc             func(session *etw.Session) error
}

func (m *mockSessionOperator) newSession(config config) (*etw.Session, error) {
	if m.newSessionFunc != nil {
		return m.newSessionFunc(config)
	}
	return nil, nil
}

func (m *mockSessionOperator) attachToExistingSession(session *etw.Session) error {
	if m.attachToExistingSessionFunc != nil {
		return m.attachToExistingSessionFunc(session)
	}
	return nil
}

func (m *mockSessionOperator) createRealtimeSession(session *etw.Session) error {
	if m.createRealtimeSessionFunc != nil {
		return m.createRealtimeSessionFunc(session)
	}
	return nil
}

func (m *mockSessionOperator) startConsumer(session *etw.Session) error {
	if m.startConsumerFunc != nil {
		return m.startConsumerFunc(session)
	}
	return nil
}

func (m *mockSessionOperator) stopSession(session *etw.Session) error {
	if m.stopSessionFunc != nil {
		return m.stopSessionFunc(session)
	}
	return nil
}

func Test_RunEtwInput_NewSessionError(t *testing.T) {
	// Mocks
	mockOperator := &mockSessionOperator{}

	// Setup the mock behavior for NewSession
	mockOperator.newSessionFunc = func(config config) (*etw.Session, error) {
		return nil, fmt.Errorf("failed creating session '%s'", config.SessionName)
	}

	// Setup input
	inputCtx := input.Context{
		Cancelation: nil,
		Logger:      logp.NewLogger("test"),
	}

	etwInput := &etwInput{
		config: config{
			ProviderName:    "Microsoft-Windows-Provider",
			SessionName:     "MySession",
			TraceLevel:      "verbose",
			MatchAnyKeyword: 0xffffffffffffffff,
			MatchAllKeyword: 0,
		},
		operator: mockOperator,
		metrics:  newInputMetrics("", ""),
	}

	// Run test
	err := etwInput.Run(inputCtx, nil)
	assert.EqualError(t, err, "error initializing ETW session: failed creating session 'MySession'")
}

func Test_RunEtwInput_AttachToExistingSessionError(t *testing.T) {
	// Mocks
	mockOperator := &mockSessionOperator{}

	// Setup the mock behavior for NewSession
	mockOperator.newSessionFunc = func(config config) (*etw.Session, error) {
		mockSession := &etw.Session{
			Name:       "MySession",
			Realtime:   true,
			NewSession: false,
		}
		return mockSession, nil
	}
	// Setup the mock behavior for AttachToExistingSession
	mockOperator.attachToExistingSessionFunc = func(session *etw.Session) error {
		return fmt.Errorf("mock error")
	}

	// Setup input
	inputCtx := input.Context{
		Cancelation: nil,
		Logger:      logp.NewLogger("test"),
	}

	etwInput := &etwInput{
		config: config{
			ProviderName:    "Microsoft-Windows-Provider",
			SessionName:     "MySession",
			TraceLevel:      "verbose",
			MatchAnyKeyword: 0xffffffffffffffff,
			MatchAllKeyword: 0,
		},
		operator: mockOperator,
		metrics:  newInputMetrics("", ""),
	}

	// Run test
	err := etwInput.Run(inputCtx, nil)
	assert.EqualError(t, err, "unable to retrieve handler: mock error")
}

func Test_RunEtwInput_CreateRealtimeSessionError(t *testing.T) {
	// Mocks
	mockOperator := &mockSessionOperator{}

	// Setup the mock behavior for NewSession
	mockOperator.newSessionFunc = func(config config) (*etw.Session, error) {
		mockSession := &etw.Session{
			Name:       "MySession",
			Realtime:   true,
			NewSession: true,
		}
		return mockSession, nil
	}
	// Setup the mock behavior for AttachToExistingSession
	mockOperator.attachToExistingSessionFunc = func(session *etw.Session) error {
		return nil
	}
	// Setup the mock behavior for CreateRealtimeSession
	mockOperator.createRealtimeSessionFunc = func(session *etw.Session) error {
		return fmt.Errorf("mock error")
	}

	// Setup input
	inputCtx := input.Context{
		Cancelation: nil,
		Logger:      logp.NewLogger("test"),
	}

	etwInput := &etwInput{
		config: config{
			ProviderName:    "Microsoft-Windows-Provider",
			SessionName:     "MySession",
			TraceLevel:      "verbose",
			MatchAnyKeyword: 0xffffffffffffffff,
			MatchAllKeyword: 0,
		},
		operator: mockOperator,
		metrics:  newInputMetrics("", ""),
	}

	// Run test
	err := etwInput.Run(inputCtx, nil)
	assert.EqualError(t, err, "realtime session could not be created: mock error")
}

func Test_RunEtwInput_StartConsumerError(t *testing.T) {
	// Mocks
	mockOperator := &mockSessionOperator{}

	// Setup the mock behavior for NewSession
	mockOperator.newSessionFunc = func(config config) (*etw.Session, error) {
		mockSession := &etw.Session{
			Name:       "MySession",
			Realtime:   true,
			NewSession: true,
		}
		return mockSession, nil
	}
	// Setup the mock behavior for AttachToExistingSession
	mockOperator.attachToExistingSessionFunc = func(session *etw.Session) error {
		return nil
	}
	// Setup the mock behavior for CreateRealtimeSession
	mockOperator.createRealtimeSessionFunc = func(session *etw.Session) error {
		return nil
	}
	// Setup the mock behavior for StartConsumer
	mockOperator.startConsumerFunc = func(session *etw.Session) error {
		return fmt.Errorf("mock error")
	}
	// Setup the mock behavior for StopSession
	mockOperator.stopSessionFunc = func(session *etw.Session) error {
		return nil
	}

	// Setup cancellation
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Setup input
	inputCtx := input.Context{
		Cancelation: ctx,
		Logger:      logp.NewLogger("test"),
	}

	etwInput := &etwInput{
		config: config{
			ProviderName:    "Microsoft-Windows-Provider",
			SessionName:     "MySession",
			TraceLevel:      "verbose",
			MatchAnyKeyword: 0xffffffffffffffff,
			MatchAllKeyword: 0,
		},
		operator: mockOperator,
		metrics:  newInputMetrics("", ""),
	}

	// Run test
	err := etwInput.Run(inputCtx, nil)
	assert.EqualError(t, err, "failed running ETW consumer: mock error")
}

func Test_RunEtwInput_Success(t *testing.T) {
	// Mocks
	mockOperator := &mockSessionOperator{}

	// Setup the mock behavior for NewSession
	mockOperator.newSessionFunc = func(config config) (*etw.Session, error) {
		mockSession := &etw.Session{
			Name:       "MySession",
			Realtime:   true,
			NewSession: true,
		}
		return mockSession, nil
	}
	// Setup the mock behavior for AttachToExistingSession
	mockOperator.attachToExistingSessionFunc = func(session *etw.Session) error {
		return nil
	}
	// Setup the mock behavior for CreateRealtimeSession
	mockOperator.createRealtimeSessionFunc = func(session *etw.Session) error {
		return nil
	}
	// Setup the mock behavior for StartConsumer
	mockOperator.startConsumerFunc = func(session *etw.Session) error {
		return nil
	}
	// Setup the mock behavior for StopSession
	mockOperator.stopSessionFunc = func(session *etw.Session) error {
		return nil
	}

	// Setup cancellation
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Setup input
	inputCtx := input.Context{
		Cancelation: ctx,
		Logger:      logp.NewLogger("test"),
	}

	etwInput := &etwInput{
		config: config{
			ProviderName:    "Microsoft-Windows-Provider",
			SessionName:     "MySession",
			TraceLevel:      "verbose",
			MatchAnyKeyword: 0xffffffffffffffff,
			MatchAllKeyword: 0,
		},
		operator: mockOperator,
		metrics:  newInputMetrics("", ""),
	}

	// Run test
	go func() {
		err := etwInput.Run(inputCtx, nil)
		if err != nil {
			t.Errorf("Run() error = %v, wantErr %v", err, false)
		}
	}()

	// Simulate waiting for a condition
	time.Sleep(time.Millisecond * 100)
	cancelFunc() // Trigger cancellation to test cleanup and goroutine exit
}

func Test_buildEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    etw.RenderedEtwEvent
		header   etw.EventHeader
		session  *etw.Session
		cfg      config
		expected mapstr.M
	}{
		{
			name: "TestStandardData",
			event: etw.RenderedEtwEvent{
				ProviderGUID: windows.GUID{
					Data1: 0x12345678,
					Data2: 0x1234,
					Data3: 0x1234,
					Data4: [8]byte{0x12, 0x34, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc},
				},
				ProcessID:   60,
				Opcode:      "foo",
				OpcodeRaw:   50,
				Keywords:    []string{"keyword1", "keyword2"},
				KeywordsRaw: 40,
				TaskRaw:     70,
				Task:        "TestTask",
				ThreadID:    80,
				Version:     90,
				Level:       "Critical",
				LevelRaw:    1, // Critical
				Channel:     "TestChannel",
				Properties: []etw.RenderedProperty{
					{
						Name:  "key",
						Value: "value",
					},
				},
			},
			header: etw.EventHeader{
				Size:          0,
				HeaderType:    0,
				Flags:         30,
				EventProperty: 30,
				TimeStamp:     133516441890350000,
				ProviderId: windows.GUID{
					Data1: 0x12345678,
					Data2: 0x1234,
					Data3: 0x1234,
					Data4: [8]byte{0x12, 0x34, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc},
				},
				EventDescriptor: etw.EventDescriptor{
					Id:      20,
					Channel: 10,
					Level:   1, // Critical
					Opcode:  50,
					Keyword: 40,
				},
				Time: 0,
				ActivityId: windows.GUID{
					Data1: 0x12345678,
					Data2: 0x1234,
					Data3: 0x1234,
					Data4: [8]byte{0x12, 0x34, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc},
				},
			},
			session: &etw.Session{
				Name: "Elastic-TestProvider",
			},
			cfg: config{
				ProviderName: "TestProvider",
			},
			expected: mapstr.M{
				"winlog": map[string]any{
					"activity_id": "{12345678-1234-1234-1234-123456789ABC}",
					"channel":     "TestChannel",
					"event_data": mapstr.M{
						"key": "value",
					},
					"flags":         []string{"NO_CPUTIME", "PRIVATE_SESSION", "STRING_ONLY", "TRACE_MESSAGE"},
					"flags_raw":     "0x1E",
					"keywords":      []string{"keyword1", "keyword2"},
					"keywords_raw":  "0x28",
					"opcode":        "foo",
					"opcode_raw":    "0x32",
					"process_id":    "60",
					"provider_guid": "{12345678-1234-1234-1234-123456789ABC}",
					"session":       "Elastic-TestProvider",
					"task":          "TestTask",
					"task_raw":      uint16(70),
					"thread_id":     "80",
					"version":       "90",
				},
				"event.code":     "20",
				"event.provider": "TestProvider",
				"event.severity": uint8(1),
				"log.level":      "critical",
			},
		},
		{
			// This case tests an unmapped severity, empty provider GUID and including logfile
			name: "TestAlternativeMetadata",
			event: etw.RenderedEtwEvent{
				ProviderGUID: windows.GUID{
					Data1: 0x12345678,
					Data2: 0x1234,
					Data3: 0x1234,
					Data4: [8]byte{0x12, 0x34, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc},
				},
				ProcessID:   60,
				Opcode:      "foo",
				OpcodeRaw:   50,
				Keywords:    []string{"keyword1", "keyword2"},
				KeywordsRaw: 40,
				TaskRaw:     70,
				Task:        "TestTask",
				ThreadID:    80,
				Version:     90,
				Channel:     "TestChannel",
				Properties: []etw.RenderedProperty{
					{
						Name:  "key",
						Value: "value",
					},
				},
			},
			header: etw.EventHeader{
				Size:          0,
				HeaderType:    0,
				Flags:         30,
				EventProperty: 30,
				TimeStamp:     133516441890350000,
				EventDescriptor: etw.EventDescriptor{
					Id:      20,
					Channel: 10,
					Opcode:  50,
					Keyword: 40,
				},
				Time: 0,
				ActivityId: windows.GUID{
					Data1: 0x12345678,
					Data2: 0x1234,
					Data3: 0x1234,
					Data4: [8]byte{0x12, 0x34, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc},
				},
			},
			session: &etw.Session{
				Name: "Elastic-TestProvider",
			},
			cfg: config{
				ProviderName: "TestProvider",
				Logfile:      "C:\\TestFile",
			},

			expected: mapstr.M{
				"winlog": map[string]any{
					"activity_id": "{12345678-1234-1234-1234-123456789ABC}",
					"channel":     "TestChannel",
					"event_data": mapstr.M{
						"key": "value",
					},
					"flags":         []string{"NO_CPUTIME", "PRIVATE_SESSION", "STRING_ONLY", "TRACE_MESSAGE"},
					"flags_raw":     "0x1E",
					"keywords":      []string{"keyword1", "keyword2"},
					"keywords_raw":  "0x28",
					"opcode":        "foo",
					"opcode_raw":    "0x32",
					"process_id":    "60",
					"provider_guid": "{12345678-1234-1234-1234-123456789ABC}",
					"session":       "Elastic-TestProvider",
					"task":          "TestTask",
					"task_raw":      uint16(70),
					"thread_id":     "80",
					"version":       "90",
				},
				"event.code":     "20",
				"event.provider": "TestProvider",
				"event.severity": uint8(0),
				"log.file.path":  "C:\\TestFile",
				"log.level":      "information",
			},
		},
	}

	for _, tt := range tests {
		//nolint:errcheck // Ignore error checks for simplicity in this test
		t.Run(tt.name, func(t *testing.T) {
			evt := buildEvent(tt.event, tt.header, tt.session, tt.cfg)
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["activity_id"], evt.Fields["winlog"].(map[string]any)["activity_id"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["channel"], evt.Fields["winlog"].(map[string]any)["channel"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["event_data"], evt.Fields["winlog"].(map[string]any)["event_data"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["flags"], evt.Fields["winlog"].(map[string]any)["flags"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["flags_raw"], evt.Fields["winlog"].(map[string]any)["flags_raw"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["keywords"], evt.Fields["winlog"].(map[string]any)["keywords"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["keywords_raw"], evt.Fields["winlog"].(map[string]any)["keywords_raw"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["opcode"], evt.Fields["winlog"].(map[string]any)["opcode"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["process_id"], evt.Fields["winlog"].(map[string]any)["process_id"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["provider_guid"], evt.Fields["winlog"].(map[string]any)["provider_guid"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["session"], evt.Fields["winlog"].(map[string]any)["session"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["task"], evt.Fields["winlog"].(map[string]any)["task"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["task_raw"], evt.Fields["winlog"].(map[string]any)["task_raw"])
			assert.Equal(t, tt.expected["winlog"].(map[string]any)["thread_id"], evt.Fields["winlog"].(map[string]any)["thread_id"])
			mapEv := evt.Fields.Flatten()

			assert.Equal(t, tt.expected["winlog"].(map[string]any)["version"], strconv.Itoa(int(mapEv["winlog.version"].(uint8))))
			assert.Equal(t, tt.expected["event.code"], mapEv["event.code"])
			assert.Equal(t, tt.expected["event.provider"], mapEv["event.provider"])
			assert.Equal(t, tt.expected["event.severity"], mapEv["event.severity"])
			assert.Equal(t, tt.expected["log.file.path"], mapEv["log.file.path"])
			assert.Equal(t, tt.expected["log.level"], mapEv["log.level"])
		})
	}
}
