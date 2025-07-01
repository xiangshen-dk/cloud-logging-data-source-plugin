// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"context"
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/GoogleCloudPlatform/cloud-logging-data-source-plugin/pkg/plugin/cloudlogging"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MockAPI struct {
	mock.Mock
}

func (m *MockAPI) ListLogs(ctx context.Context, query *cloudlogging.Query) ([]*loggingpb.LogEntry, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*loggingpb.LogEntry), args.Error(1)
}

func (m *MockAPI) TestConnection(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockAPI) ListProjects(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAPI) ListProjectBuckets(ctx context.Context, projectId string) ([]string, error) {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAPI) ListProjectBucketViews(ctx context.Context, projectId string, bucketId string) ([]string, error) {
	args := m.Called(ctx, projectId, bucketId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAPI) Close() error {
	args := m.Called()
	return args.Error(0)
}

// This is where the tests for the datasource backend live.
func TestQueryData(t *testing.T) {
	ds := CloudLoggingDatasource{}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A"},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}

func TestQueryData_InvalidJSON(t *testing.T) {
	client := &MockAPI{}
	ds := CloudLoggingDatasource{
		client: client,
	}
	refID := "test"
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			{
				JSON:  []byte(`Not JSON`),
				RefID: refID,
			},
		},
	})

	require.NoError(t, err)
	require.Error(t, resp.Responses[refID].Error)
	require.Nil(t, resp.Responses[refID].Frames)
	client.AssertExpectations(t)
}

func TestQueryData_GCPError(t *testing.T) {
	to := time.Now()
	from := to.Add(-1 * time.Hour)
	expectedErr := errors.New("something was wrong with the request")

	client := &MockAPI{}
	client.On("ListLogs", mock.Anything, &cloudlogging.Query{
		ProjectID: "testing",
		Filter:    `resource.type = "testing"`,
		Limit:     20,
		TimeRange: struct {
			From string
			To   string
		}{
			From: from.Format(time.RFC3339),
			To:   to.Format(time.RFC3339),
		},
	}).Return(nil, expectedErr)

	ds := CloudLoggingDatasource{
		client: client,
	}
	refID := "test"
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			{
				JSON:  []byte(`{"projectId": "testing", "queryText": "resource.type = \"testing\""}`),
				RefID: refID,
				TimeRange: backend.TimeRange{
					From: from,
					To:   to,
				},
				MaxDataPoints: 20,
			},
		},
	})

	require.NoError(t, err)
	require.ErrorContains(t, resp.Responses[refID].Error, expectedErr.Error())
	require.Nil(t, resp.Responses[refID].Frames)
	client.AssertExpectations(t)
}

func TestQueryData_SingleLog(t *testing.T) {
	to := time.Now()
	from := to.Add(-1 * time.Hour)
	// insertID and receivedAt are hardcoded to match the expected response
	insertID := "b6f39be2-b298-44da-9001-1f04e5756fa0"
	receivedAt := timestamppb.New(time.UnixMilli(1660920349373))
	trace := "projects/xxx/traces/c0e331eab1515bbcd1b8306029902ff7"

	logEntry := loggingpb.LogEntry{
		LogName: "organizations/1234567890/logs/cloudresourcemanager.googleapis.com%2Factivity",
		Resource: &monitoredres.MonitoredResource{
			Type:   "gce_instance",
			Labels: map[string]string{},
		},
		Timestamp:        receivedAt,
		ReceiveTimestamp: receivedAt,
		Severity:         ltype.LogSeverity_INFO,
		InsertId:         insertID,
		Trace:            trace,
		Labels: map[string]string{
			"instance_id":  "unique",
			"custom_label": "custom_value",
		},
		Payload: &loggingpb.LogEntry_TextPayload{
			TextPayload: "Full log message from this GCE instance",
		},
	}

	client := &MockAPI{}
	client.On("ListLogs", mock.Anything, &cloudlogging.Query{
		ProjectID: "testing",
		Filter:    `resource.type = "testing"`,
		Limit:     20,
		TimeRange: struct {
			From string
			To   string
		}{
			From: from.Format(time.RFC3339),
			To:   to.Format(time.RFC3339),
		},
	}).Return([]*loggingpb.LogEntry{&logEntry}, nil)
	client.On("Close").Return(nil)

	ds := CloudLoggingDatasource{
		client: client,
	}
	refID := "test"
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			{
				JSON:  []byte(`{"projectId": "testing", "queryText": "resource.type = \"testing\""}`),
				RefID: refID,
				TimeRange: backend.TimeRange{
					From: from,
					To:   to,
				},
				MaxDataPoints: 20,
			},
		},
	})
	ds.Dispose()
	require.NoError(t, err)
	require.Len(t, resp.Responses[refID].Frames, 1)

	frame := resp.Responses[refID].Frames[0]
	require.Equal(t, insertID, frame.Name)
	require.Len(t, frame.Fields, 2)
	require.Equal(t, data.VisTypeLogs, string(frame.Meta.PreferredVisualization))

	expectedFrame := []byte(`{"schema":{"name":"b6f39be2-b298-44da-9001-1f04e5756fa0","meta":{"typeVersion":[0,0],"preferredVisualisationType":"logs"},"fields":[{"name":"time","type":"time","typeInfo":{"frame":"time.Time"}},{"name":"content","type":"string","typeInfo":{"frame":"string"},"labels":{"id":"b6f39be2-b298-44da-9001-1f04e5756fa0","labels.\"custom_label\"":"custom_value","labels.\"instance_id\"":"unique","level":"info","resource.type":"gce_instance","textPayload":"Full log message from this GCE instance","trace":"projects/xxx/traces/c0e331eab1515bbcd1b8306029902ff7","traceId":"c0e331eab1515bbcd1b8306029902ff7"}}]},"data":{"values":[[1660920349373],["Full log message from this GCE instance"]]}}`)

	serializedFrame, err := frame.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(expectedFrame), string(serializedFrame))
	client.AssertExpectations(t)
}

func TestCallResource(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		projectErr     error
		bucketErr      error
		viewErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "projects endpoint success",
			resource:       "projects",
			projectErr:     nil,
			expectedStatus: 200,
			expectedBody:   `["project1","project2"]`,
		},
		{
			name:           "projects endpoint failure",
			resource:       "projects",
			projectErr:     errors.New("permission denied"),
			expectedStatus: 502,
			expectedBody:   `{"error": "Failed to list projects. Please check your permissions and authentication configuration."}`,
		},
		{
			name:           "logbuckets endpoint success",
			resource:       "logbuckets",
			bucketErr:      nil,
			expectedStatus: 200,
			expectedBody:   `["bucket1","bucket2"]`,
		},
		{
			name:           "logbuckets endpoint failure",
			resource:       "logbuckets",
			bucketErr:      errors.New("invalid project"),
			expectedStatus: 502,
			expectedBody:   `{"error": "Failed to list log buckets. Please check your project ID and permissions."}`,
		},
		{
			name:           "logviews endpoint success",
			resource:       "logviews",
			viewErr:        nil,
			expectedStatus: 200,
			expectedBody:   `["view1","view2"]`,
		},
		{
			name:           "logviews endpoint failure",
			resource:       "logviews",
			viewErr:        errors.New("invalid bucket"),
			expectedStatus: 502,
			expectedBody:   `{"error": "Failed to list log views. Please check your bucket ID and permissions."}`,
		},
		{
			name:           "unknown resource",
			resource:       "unknown",
			expectedStatus: 404,
			expectedBody:   `No such path`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockAPI{}

			// Set up mocks based on test case
			if tt.resource == "projects" {
				if tt.projectErr != nil {
					client.On("ListProjects", mock.Anything).Return(nil, tt.projectErr)
				} else {
					client.On("ListProjects", mock.Anything).Return([]string{"project1", "project2"}, nil)
				}
			} else if tt.resource == "logbuckets" {
				if tt.bucketErr != nil {
					client.On("ListProjectBuckets", mock.Anything, mock.Anything).Return(nil, tt.bucketErr)
				} else {
					client.On("ListProjectBuckets", mock.Anything, mock.Anything).Return([]string{"bucket1", "bucket2"}, nil)
				}
			} else if tt.resource == "logviews" {
				if tt.viewErr != nil {
					client.On("ListProjectBucketViews", mock.Anything, mock.Anything, mock.Anything).Return(nil, tt.viewErr)
				} else {
					client.On("ListProjectBucketViews", mock.Anything, mock.Anything, mock.Anything).Return([]string{"view1", "view2"}, nil)
				}
			}

			ds := &CloudLoggingDatasource{
				client: client,
			}

			// Create a mock sender to capture the response
			sender := &mockCallResourceResponseSender{
				response: &backend.CallResourceResponse{},
			}

			req := &backend.CallResourceRequest{
				Path: tt.resource,
				URL:  "http://localhost/" + tt.resource,
			}

			if tt.resource == "logbuckets" || tt.resource == "logviews" {
				req.URL += "?ProjectId=test-project&BucketId=test-bucket"
			}

			err := ds.CallResource(context.Background(), req, sender)
			require.NoError(t, err)

			require.Equal(t, tt.expectedStatus, sender.response.Status)
			require.Equal(t, tt.expectedBody, string(sender.response.Body))

			if tt.resource != "unknown" {
				client.AssertExpectations(t)
			}
		})
	}
}

// mockCallResourceResponseSender implements backend.CallResourceResponseSender
type mockCallResourceResponseSender struct {
	response *backend.CallResourceResponse
}

func (m *mockCallResourceResponseSender) Send(response *backend.CallResourceResponse) error {
	m.response = response
	return nil
}
