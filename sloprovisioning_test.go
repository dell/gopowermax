/*
 *
 * Copyright © 2021-2026 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pmax

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/dell/gopowermax/v2/mock"
	types "github.com/dell/gopowermax/v2/types/v100"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPortListByProtocol(t *testing.T) {
	allowdArray := "testSymID"
	tests := []struct {
		name           string
		symID          string
		protocol       string
		expectedPorts  *types.PortList
		expectedStatus int
		expectedErr    string
	}{
		{
			name:           "Valid symID and protocol",
			symID:          "testSymID",
			protocol:       "SCSI_FC",
			expectedPorts:  &types.PortList{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty protocol",
			symID:          "testSymID",
			protocol:       "",
			expectedPorts:  &types.PortList{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error from API",
			symID:          "testSymID",
			protocol:       "SCSI_FC",
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    "500 Internal Server Error",
		},
		{
			name:        "Error from IsAllowedArray",
			symID:       "",
			protocol:    "SCSI_FC",
			expectedErr: "the requested array () is ignored as it is not managed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/univmax/restapi/100/sloprovisioning/symmetrix/"+tc.symID+"/port", r.URL.Path, "Expected URL")
				assert.Equal(t, tc.protocol, r.URL.Query().Get("enabled_protocol"), "Expected enabled_protocol")
				w.WriteHeader(tc.expectedStatus)
				json.NewEncoder(w).Encode(tc.expectedPorts)
			}))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowdArray})

			portList, err := c.GetPortListByProtocol(context.Background(), tc.symID, tc.protocol)
			assert.Equal(t, tc.expectedPorts, portList)
			if tc.expectedStatus != http.StatusOK {
				assert.ErrorContains(t, err, tc.expectedErr)
			}
		})
	}
}

func TestGetVolumesByIdentifier(t *testing.T) {
	allowedArray := "testSymID"
	volID := "vol-0001"
	tests := []struct {
		name            string
		symID           string
		expectedVolumes *types.Volumev1
		expectedStatus  int
		expectedErr     string
	}{
		{
			name:  "Valid symID with volumes",
			symID: "testSymID",
			expectedVolumes: &types.Volumev1{
				Volumes: []types.VolumeEnhanced{
					{
						ID:         "0001",
						Type:       "TD",
						System:     types.SystemInfo{ID: "testSymID"},
						Identifier: "vol-0001",
						StorageGroups: []types.StorageGroupID{
							{StorageGroupID: "SG1"},
						},
						CapCyl: 1000,
						MaskingViews: []types.MaskingViewID{
							{MaskingViewID: "MV1"},
						},
						VolumeHostPaths:      []types.VolumeHostPath{{ID: "1"}},
						NumberOfMaskingViews: 1,
						SRP:                  types.Srp{ID: "1"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "Valid symID with no volumes",
			symID:           "testSymID",
			expectedVolumes: &types.Volumev1{Volumes: []types.VolumeEnhanced{}},
			expectedStatus:  http.StatusOK,
		},
		{
			name:        "Error from IsAllowedArray",
			symID:       "",
			expectedErr: "the requested array () is ignored as it is not managed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/univmax/rest/v1/systems/"+tc.symID+"/volumes", r.URL.Path, "Expected URL")
				w.WriteHeader(tc.expectedStatus)
				if tc.expectedStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tc.expectedVolumes)
				}
			}))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			volumes, err := c.GetVolumesByIdentifier(context.Background(), tc.symID, volID)
			assert.Equal(t, tc.expectedVolumes, volumes)
			if tc.expectedStatus != http.StatusOK {
				assert.ErrorContains(t, err, tc.expectedErr)
			}
		})
	}
}

func TestGetVolumesByIdentifierMatch(t *testing.T) {
	allowedArray := "testSymID"
	identifierMatcher := "vol-.*"
	tests := []struct {
		name            string
		symID           string
		expectedVolumes *types.Volumev1
		expectedStatus  int
		expectedErr     string
		responseBody    string
	}{
		{
			name:  "Valid symID with matching volumes",
			symID: "testSymID",
			expectedVolumes: &types.Volumev1{
				Volumes: []types.VolumeEnhanced{
					{
						ID:         "0001",
						Type:       "TD",
						System:     types.SystemInfo{ID: "testSymID"},
						Identifier: "vol-0001",
						StorageGroups: []types.StorageGroupID{
							{StorageGroupID: "SG1"},
						},
						CapCyl: 1000,
						MaskingViews: []types.MaskingViewID{
							{MaskingViewID: "MV1"},
						},
						VolumeHostPaths:      []types.VolumeHostPath{{ID: "1"}},
						NumberOfMaskingViews: 1,
						SRP:                  types.Srp{ID: "1"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "Valid symID with no volumes returns empty slice",
			symID:           "testSymID",
			expectedVolumes: &types.Volumev1{Volumes: []types.VolumeEnhanced{}},
			expectedStatus:  http.StatusOK,
		},
		{
			name:        "Error from IsAllowedArray",
			symID:       "",
			expectedErr: "the requested array () is ignored as it is not managed",
		},
		{
			name:           "HTTP error from API",
			symID:          "testSymID",
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    "Internal Server Error",
		},
		{
			name:           "Invalid JSON response",
			symID:          "testSymID",
			expectedStatus: http.StatusOK,
			responseBody:   "not-valid-json",
			expectedErr:    "invalid character",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/univmax/rest/v1/systems/"+tc.symID+"/volumes", r.URL.Path, "Expected URL path")
				assert.Contains(t, r.URL.RawQuery, "filter=identifier%20like%20"+identifierMatcher, "Expected like filter in query")
				assert.Contains(t, r.URL.RawQuery, "limit=100", "Expected limit param")
				assert.Contains(t, r.URL.RawQuery, "expiration_delay_secs=30", "Expected expiration_delay_secs param")
				w.WriteHeader(tc.expectedStatus)
				if tc.responseBody != "" {
					w.Write([]byte(tc.responseBody))
				} else if tc.expectedStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tc.expectedVolumes)
				}
			}))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			volumes, err := c.GetVolumesByIdentifierMatch(context.Background(), tc.symID, identifierMatcher)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, volumes)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVolumes, volumes)
			}
		})
	}
}

func TestGetVolumesByIdentifierMatchPagination(t *testing.T) {
	allowedArray := "testSymID"
	identifierMatcher := "vol-.*"
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		switch requestCount {
		case 1:
			// First page: return a volume and a resume token
			assert.NotContains(t, r.URL.RawQuery, "resume_token", "First request should not have resume_token")
			json.NewEncoder(w).Encode(types.Volumev1{
				Volumes: []types.VolumeEnhanced{
					{ID: "0001", Identifier: "vol-0001"},
				},
				VolumePaging: types.VolumePaging{ResumeToken: "token-page-2", RemainingInstances: 2},
			})
		case 2:
			// Second page: return another volume and a resume token
			assert.Contains(t, r.URL.RawQuery, "resume_token=token-page-2", "Second request should have resume_token")
			json.NewEncoder(w).Encode(types.Volumev1{
				Volumes: []types.VolumeEnhanced{
					{ID: "0002", Identifier: "vol-0002"},
				},
				VolumePaging: types.VolumePaging{ResumeToken: "token-page-3", RemainingInstances: 1},
			})
		case 3:
			// Last page: return a volume with no resume token
			assert.Contains(t, r.URL.RawQuery, "resume_token=token-page-3", "Third request should have resume_token")
			json.NewEncoder(w).Encode(types.Volumev1{
				Volumes: []types.VolumeEnhanced{
					{ID: "0003", Identifier: "vol-0003"},
				},
			})
		}
	}))
	defer server.Close()

	c, err := NewClientWithArgs(server.URL, "", true, true, "")
	assert.NoError(t, err)
	c.SetAllowedArrays([]string{allowedArray})

	volumes, err := c.GetVolumesByIdentifierMatch(context.Background(), "testSymID", identifierMatcher)
	assert.NoError(t, err)
	assert.Equal(t, 3, requestCount, "Expected 3 paginated requests")
	assert.Equal(t, 3, len(volumes.Volumes), "Expected 3 volumes across all pages")
	assert.Equal(t, "0001", volumes.Volumes[0].ID)
	assert.Equal(t, "0002", volumes.Volumes[1].ID)
	assert.Equal(t, "0003", volumes.Volumes[2].ID)
}

func TestGetVolumesCapacityBulk(t *testing.T) {
	allowedArray := "testSymID"
	tests := []struct {
		name            string
		symID           string
		expectedVolumes *types.Volumev1
		expectedStatus  int
		expectedErr     string
		responseBody    string
	}{
		{
			name:  "Valid symID with volumes",
			symID: "testSymID",
			expectedVolumes: &types.Volumev1{
				Volumes: []types.VolumeEnhanced{
					{
						ID:                      "0001",
						CapGB:                   100,
						EffectiveUsedCapacityGB: 25,
						StorageGroups: []types.StorageGroupID{
							{StorageGroupID: "SG1"},
						},
						SRP: types.Srp{ID: "SRP_1"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "Valid symID with no volumes returns empty slice",
			symID:           "testSymID",
			expectedVolumes: &types.Volumev1{Volumes: []types.VolumeEnhanced{}},
			expectedStatus:  http.StatusOK,
		},
		{
			name:        "Error from IsAllowedArray",
			symID:       "",
			expectedErr: "the requested array () is ignored as it is not managed",
		},
		{
			name:           "HTTP error from API",
			symID:          "testSymID",
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    "Internal Server Error",
		},
		{
			name:           "Invalid JSON response",
			symID:          "testSymID",
			expectedStatus: http.StatusOK,
			responseBody:   "not-valid-json",
			expectedErr:    "invalid character",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/univmax/rest/v1/systems/"+tc.symID+"/volumes", r.URL.Path, "Expected URL path")
				assert.Contains(t, r.URL.RawQuery, "select=id,cap_gb,effective_used_capacity_gb", "Expected capacity select fields")
				assert.Contains(t, r.URL.RawQuery, "limit=1000", "Expected limit param")
				assert.Contains(t, r.URL.RawQuery, "expiration_delay_secs=30", "Expected expiration_delay_secs param")
				w.WriteHeader(tc.expectedStatus)
				if tc.responseBody != "" {
					w.Write([]byte(tc.responseBody))
				} else if tc.expectedStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tc.expectedVolumes)
				}
			}))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			volumes, err := c.GetVolumesCapacityBulk(context.Background(), tc.symID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, volumes)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVolumes, volumes)
			}
		})
	}
}

func TestGetVolumesCapacityBulkPagination(t *testing.T) {
	allowedArray := "testSymID"
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		switch requestCount {
		case 1:
			assert.NotContains(t, r.URL.RawQuery, "resume_token", "First request should not have resume_token")
			json.NewEncoder(w).Encode(types.Volumev1{
				Volumes:      []types.VolumeEnhanced{{ID: "0001", CapGB: 10}},
				VolumePaging: types.VolumePaging{ResumeToken: "token-page-2", RemainingInstances: 1},
			})
		case 2:
			assert.Contains(t, r.URL.RawQuery, "resume_token=token-page-2", "Second request should have resume_token")
			json.NewEncoder(w).Encode(types.Volumev1{
				Volumes: []types.VolumeEnhanced{{ID: "0002", CapGB: 20}},
			})
		}
	}))
	defer server.Close()

	c, err := NewClientWithArgs(server.URL, "", true, true, "")
	assert.NoError(t, err)
	c.SetAllowedArrays([]string{allowedArray})

	volumes, err := c.GetVolumesCapacityBulk(context.Background(), "testSymID")
	assert.NoError(t, err)
	assert.Equal(t, 2, requestCount, "Expected 2 paginated requests")
	assert.Equal(t, 2, len(volumes.Volumes), "Expected 2 volumes across all pages")
	assert.Equal(t, "0001", volumes.Volumes[0].ID)
	assert.Equal(t, "0002", volumes.Volumes[1].ID)
}

func TestStorageGroupVolumeCounts(t *testing.T) {
	allowedArray := "testSymID"
	tests := []struct {
		name       string
		symID      string
		url        string
		httpStatus int
		httpBody   string
		want       *types.StorageGroupVolumeCounts
		wantErr    bool
	}{
		{
			name:       "valid response",
			symID:      "testSymID",
			url:        "/univmax/rest/v1/systems/testSymID/storage-groups",
			httpStatus: http.StatusOK,
			httpBody: `{
				"storage_groups": [
					{
						"id": "csi-rep-sg-repctl-0908-t083859-123-SYNC",
						"num_of_volumes": 1
					},
					{
						"id": "csi-dk0-dk-Bronze-SRP_1-SG-myhostlimit",
						"num_of_volumes": 0
					},
					{
						"id": "csi-rep-sg-repctl-0909-t141146-121-SYNC",
						"num_of_volumes": 1
					}
				]
			}`,
			want: &types.StorageGroupVolumeCounts{
				StorageGroups: []types.StorageGroupVolumeCount{
					{
						ID:          "csi-rep-sg-repctl-0908-t083859-123-SYNC",
						VolumeCount: 1,
					},
					{
						ID:          "csi-dk0-dk-Bronze-SRP_1-SG-myhostlimit",
						VolumeCount: 0,
					},
					{
						ID:          "csi-rep-sg-repctl-0909-t141146-121-SYNC",
						VolumeCount: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "invalid json",
			symID:      "testSymID",
			url:        "/univmax/rest/v1/systems/testSymID/storage-groups",
			httpStatus: http.StatusOK,
			httpBody:   "invalid json",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "http error",
			symID:      "testSymID",
			url:        "/univmax/rest/v1/systems/testSymID/storage-groups",
			httpStatus: http.StatusInternalServerError,
			httpBody:   "",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "empty response",
			symID:      "testSymID",
			url:        "/univmax/rest/v1/systems/testSymID/storage-groups",
			httpStatus: http.StatusOK,
			httpBody: `{
				"storage_groups": []
			}`,
			want: &types.StorageGroupVolumeCounts{
				StorageGroups: []types.StorageGroupVolumeCount{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.url {
					t.Errorf("expected URL %s, got %s", tt.url, r.URL.Path)
				}
				w.WriteHeader(tt.httpStatus)
				w.Write([]byte(tt.httpBody))
			}))
			defer ts.Close()

			// Create a client
			c, err := NewClientWithArgs(ts.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			// Call the method
			got, err := c.GetStorageGroupVolumeCounts(context.TODO(), tt.symID, "")

			// Check the error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorageGroupVolumeCounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check the response
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStorageGroupVolumeCounts() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateVolume(t *testing.T) {
	allowedArray := "testSymID"
	systemID := "testSymID"

	tests := []struct {
		name         string
		systemID     string
		req          types.CreateVolumesRequest
		httpStatus   int
		httpBody     string
		expectedErr  string
		expectedResp *types.CreateVolumesResponse
		wantErr      bool
	}{
		{
			name:     "Successful volume creation",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "test-vol-001",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedResp: &types.CreateVolumesResponse{
				Summary: types.ResponseSummary{
					Total:              1,
					Succeeded:          1,
					Failed:             0,
					Rejected:           0,
					PartiallySucceeded: 0,
				},
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{},
				},
			},
			wantErr: false,
		},
		{
			name:     "Empty volumes request",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{},
			},
			expectedErr: "create volumes request cannot be empty",
			wantErr:     true,
		},
		{
			name:     "Invalid array",
			systemID: "invalidArray",
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "test-vol-001",
							},
						},
					},
				},
			},
			expectedErr: "is ignored as it is not managed",
			wantErr:     true,
		},
		{
			name:     "HTTP error response",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "test-vol-001",
							},
						},
					},
				},
			},
			httpStatus:  http.StatusInternalServerError,
			httpBody:    `{"summary": {"total": 0, "succeeded": 0, "failed": 0, "rejected": 0}, "results": {"result": []}}`,
			expectedErr: "500",
			wantErr:     true,
		},
		{
			name:     "Partial failure in response",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "test-vol-001",
							},
						},
					},
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   2048,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "test-vol-002",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 2,
					"succeeded": 1,
					"failed": 1,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedErr: "create volumes failed",
			wantErr:     true,
		},
		{
			name:     "No volumes succeeded",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "test-vol-001",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 0,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedErr: "none succeeded",
			wantErr:     true,
		},
		{
			name:     "Response with error messages",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "test-vol-001",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 0,
					"failed": 1,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": [
						{
							"messages": {
								"message": [
									{
										"code": "ERR001",
										"message": "Volume already exists"
									}
								]
							}
						}
					]
				}
			}`,
			expectedErr: "ERR001: Volume already exists",
			wantErr:     true,
		},
		{
			name:     "Clone volume using create_new_from_attributes with CopyFrom",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						Volume: &types.ExistingVolumeRequestParam{Identifier: "cloned-vol-001"},
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								CapacityUnit: "CYL",
								VolumeSize:   547,
							},
							PrecheckSrpCapacity: &types.ValidationSrpAction{
								SRP: types.VolumeSrpParam{ID: "SRP_1"},
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageVolumeStorageGroup: &types.ManageVolumeStorageGroupAction{
								Action:       "Add",
								StorageGroup: types.VolumeStorageGroupParam{ID: "test-SG"},
							},
							ManageReplication: &types.ManageReplicationAction{
								Local: &types.LocalReplicationAction{
									Action:             "CopyFrom",
									Volume:             types.ExistingVolumeRequestParam{ID: "0046A"},
									EstablishTerminate: func() *bool { v := true; return &v }(),
								},
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedResp: &types.CreateVolumesResponse{
				Summary: types.ResponseSummary{
					Total:              1,
					Succeeded:          1,
					Failed:             0,
					Rejected:           0,
					PartiallySucceeded: 0,
				},
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{},
				},
			},
			wantErr: false,
		},
		{
			name:     "Create volume from snapshot with new_volume_attributes",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						Volume: &types.ExistingVolumeRequestParam{Identifier: "snapshot-vol-001"},
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromSnapshot: &types.CreateNewFromSnapshot{
								Snapshot: types.SnapshotRequestParam{
									ID: "95935652609",
								},
								NewVolumeAttributes: &types.CreateNewFromAttributes{
									CapacityUnit: "CYL",
									VolumeSize:   547,
								},
							},
							PrecheckSrpCapacity: &types.ValidationSrpAction{
								SRP: types.VolumeSrpParam{ID: "SRP_1"},
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageVolumeStorageGroup: &types.ManageVolumeStorageGroupAction{
								Action:       "Add",
								StorageGroup: types.VolumeStorageGroupParam{ID: "test-SG"},
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedResp: &types.CreateVolumesResponse{
				Summary: types.ResponseSummary{
					Total:              1,
					Succeeded:          1,
					Failed:             0,
					Rejected:           0,
					PartiallySucceeded: 0,
				},
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{},
				},
			},
			wantErr: false,
		},
		{
			name:     "Create volume with SRP capacity precheck",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   2048,
								CapacityUnit: "GB",
							},
							PrecheckSrpCapacity: &types.ValidationSrpAction{
								SRP: types.VolumeSrpParam{
									ID: "SRP_1",
								},
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "precheck-vol-001",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedResp: &types.CreateVolumesResponse{
				Summary: types.ResponseSummary{
					Total:              1,
					Succeeded:          1,
					Failed:             0,
					Rejected:           0,
					PartiallySucceeded: 0,
				},
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{},
				},
			},
			wantErr: false,
		},
		{
			name:     "Complete real-world scenario with all actions",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1,
								CapacityUnit: "GB",
							},
							PrecheckSrpCapacity: &types.ValidationSrpAction{
								SRP: types.VolumeSrpParam{
									ID: "SRP_1",
								},
							},
						},
						ResponseSelect: "id,identifier,storage_groups",
						RequestID:      "csi-test-1",
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "csi-test-vol-1",
							},
							ManageVolumeStorageGroup: &types.ManageVolumeStorageGroupAction{
								Action: "Add",
								StorageGroup: types.VolumeStorageGroupParam{
									ID: "csi-CSM-Silver-SRP_1-SG",
								},
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"http_status_code": 200,
				"summary": {
					"total": 1,
					"partially_succeeded": 0,
					"succeeded": 1,
					"failed": 0,
					"not_run": 0,
					"rejected": 0
				},
				"results": {
					"result": [
						{
							"volume": {
								"id": "00187",
								"identifier": "csi-test-vol-1",
								"storage_groups": [
									{
										"id": "csi-CSM-Silver-SRP_1-SG"
									}
								]
							},
							"storage_group": {
								"id": "csi-CSM-Silver-SRP_1-SG",
								"num_of_volumes": 27
							},
							"status": "success",
							"steps": [
								{
									"status": "success",
									"description": "Create [1] volume, add volume to storage group [csi-CSM-Silver-SRP_1-SG], and set identifier to [csi-test-vol-1]",
									"result": "The following volume was created [00187], added to storage group [csi-CSM-Silver-SRP_1-SG], and Identifier set to [csi-test-vol-1]"
								}
							],
							"request_id": "csi-test-1",
							"resource_id": "Volume"
						}
					]
				}
			}`,
			expectedResp: &types.CreateVolumesResponse{
				Summary: types.ResponseSummary{
					Total:              1,
					Succeeded:          1,
					Failed:             0,
					Rejected:           0,
					PartiallySucceeded: 0,
				},
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{
						{
							Status:     "success",
							RequestID:  "csi-test-1",
							ResourceID: "Volume",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Rejected volumes in response",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "rejected-vol-001",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 0,
					"failed": 0,
					"rejected": 1,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedErr: "create volumes failed",
			wantErr:     true,
		},
		{
			name:     "Create volume with ManageReplication action",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "repl-vol-001",
							},
							ManageReplication: &types.ManageReplicationAction{
								Local: &types.LocalReplicationAction{
									Action: "CopyFrom",
									Volume: types.ExistingVolumeRequestParam{
										ID: "0046A",
									},
								},
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedResp: &types.CreateVolumesResponse{
				Summary: types.ResponseSummary{
					Total:     1,
					Succeeded: 1,
				},
			},
			wantErr: false,
		},
		{
			name:     "Error message without code",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   1024,
								CapacityUnit: "GB",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 0,
					"failed": 1,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": [
						{
							"messages": {
								"message": [
									{
										"code": "",
										"message": "Insufficient capacity on SRP"
									}
								]
							}
						}
					]
				}
			}`,
			expectedErr: "Insufficient capacity on SRP",
			wantErr:     true,
		},
		{
			name:     "Response with cap_cyl in VolumeRefResponse",
			systemID: systemID,
			req: types.CreateVolumesRequest{
				ResponseSelect: "id,identifier,cap_cyl",
				Volumes: []types.VolumeRequestParam{
					{
						CreateNew: &types.CreateVolumeParam{
							CreateNewFromAttributes: &types.CreateNewFromAttributes{
								VolumeSize:   547,
								CapacityUnit: "CYL",
							},
						},
						ResponseSelect: "id,identifier,cap_cyl",
						Actions: &types.VolumeRequestParamActions{
							ManageIdentifier: &types.ManageIdentifierAction{
								Identifier: "capcyl-vol-001",
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": [
						{
							"volume": {
								"id": "00200",
								"identifier": "capcyl-vol-001",
								"cap_cyl": 547
							},
							"status": "success",
							"request_id": "capcyl-req-1",
							"resource_id": "Volume"
						}
					]
				}
			}`,
			expectedResp: &types.CreateVolumesResponse{
				Summary: types.ResponseSummary{
					Total:     1,
					Succeeded: 1,
				},
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{
						{
							Volume: &types.VolumeRefResponse{
								ID:         "00200",
								Identifier: "capcyl-vol-001",
								CapCyl:     547,
							},
							Status:    "success",
							RequestID: "capcyl-req-1",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method, "Expected POST method")
				assert.Contains(t, r.URL.Path, "/systems/"+tc.systemID+"/volumes", "Expected URL path")
				w.WriteHeader(tc.httpStatus)
				w.Write([]byte(tc.httpBody))
			}))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			resp, err := c.CreateVolume(context.Background(), tc.systemID, tc.req)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != "" {
					assert.ErrorContains(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tc.expectedResp != nil {
					assert.Equal(t, tc.expectedResp.Summary.Total, resp.Summary.Total)
					assert.Equal(t, tc.expectedResp.Summary.Succeeded, resp.Summary.Succeeded)
					assert.Equal(t, tc.expectedResp.Summary.Failed, resp.Summary.Failed)
					if len(tc.expectedResp.Results.Result) > 0 {
						assert.Equal(t, len(tc.expectedResp.Results.Result), len(resp.Results.Result))
						for i, expItem := range tc.expectedResp.Results.Result {
							if i >= len(resp.Results.Result) {
								break
							}
							gotItem := resp.Results.Result[i]
							assert.Equal(t, expItem.Status, gotItem.Status)
							assert.Equal(t, expItem.RequestID, gotItem.RequestID)
							if expItem.Volume != nil && gotItem.Volume != nil {
								assert.Equal(t, expItem.Volume.ID, gotItem.Volume.ID)
								assert.Equal(t, expItem.Volume.Identifier, gotItem.Volume.Identifier)
								assert.Equal(t, expItem.Volume.CapCyl, gotItem.Volume.CapCyl)
							}
						}
					}
				}
			}
		})
	}
}

func TestCreateVolumeIdempotency(t *testing.T) {
	mock.Reset()
	server := httptest.NewServer(mock.GetHandler())
	defer server.Close()

	systemID := mock.DefaultSymmetrixID
	sgID := mock.DefaultStorageGroup

	c, err := NewClientWithArgs(server.URL, "", true, true, "")
	assert.NoError(t, err)
	c.SetAllowedArrays([]string{systemID})

	req := types.CreateVolumesRequest{
		Volumes: []types.VolumeRequestParam{
			{
				Volume:    &types.ExistingVolumeRequestParam{Identifier: "idempotent-vol-001"},
				RequestID: "idempotent-vol-001",
				CreateNew: &types.CreateVolumeParam{
					CreateNewFromAttributes: &types.CreateNewFromAttributes{
						CapacityUnit: "CYL",
						VolumeSize:   547,
					},
				},
				Actions: &types.VolumeRequestParamActions{
					ManageVolumeStorageGroup: &types.ManageVolumeStorageGroupAction{
						Action:       "Add",
						StorageGroup: types.VolumeStorageGroupParam{ID: sgID},
					},
				},
				ResponseSelect: "id,identifier,cap_cyl,storage_groups",
			},
		},
	}

	resp1, err := c.CreateVolume(context.Background(), systemID, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp1)
	assert.Equal(t, 1, resp1.Summary.Succeeded)
	assert.Equal(t, 0, resp1.Summary.Failed)
	require1 := resp1.Results.Result[0]
	assert.Equal(t, "success", require1.Status)
	assert.Equal(t, "Volume", require1.ResourceID)
	assert.NotNil(t, require1.Volume)
	assert.NotEmpty(t, require1.Volume.ID)
	assert.Equal(t, "idempotent-vol-001", require1.Volume.Identifier)
	assert.Equal(t, float64(547), require1.Volume.CapCyl)

	resp2, err := c.CreateVolume(context.Background(), systemID, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp2)
	assert.Equal(t, 1, resp2.Summary.Succeeded)
	assert.Equal(t, 0, resp2.Summary.Failed)
	require2 := resp2.Results.Result[0]
	assert.Equal(t, "success", require2.Status)
	assert.Equal(t, "Volume", require2.ResourceID)
	assert.NotNil(t, require2.Volume, "idempotent response must include volume object")
	assert.Equal(t, require1.Volume.ID, require2.Volume.ID, "idempotent response must return same volume ID")
	assert.Equal(t, "idempotent-vol-001", require2.Volume.Identifier)
	assert.Equal(t, float64(547), require2.Volume.CapCyl, "idempotent response must include cap_cyl")
}

func TestCreateVolumeWithAuthHeaders(t *testing.T) {
	allowedArray := "testSymID"
	systemID := "testSymID"

	successBody := `{
		"summary": {"total": 1, "succeeded": 1, "failed": 0, "rejected": 0, "partially_succeeded": 0},
		"results": {"result": []}
	}`

	t.Run("Headers are sent on the HTTP request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the auth metadata headers are present and have correct values
			assert.Equal(t, "my-test-pvc", r.Header.Get("x-csi-pv-claimname"), "PVC claim name header")
			assert.Equal(t, "my-test-pv", r.Header.Get("x-csi-pv-name"), "PV name header")
			assert.Equal(t, "my-namespace", r.Header.Get("x-csi-pv-namespace"), "PVC namespace header")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(successBody))
		}))
		defer server.Close()

		c, err := NewClientWithArgs(server.URL, "", true, true, "")
		assert.NoError(t, err)
		c.SetAllowedArrays([]string{allowedArray})

		authHeader := http.Header{
			"x-csi-pv-claimname": []string{"my-test-pvc"},
			"x-csi-pv-name":      []string{"my-test-pv"},
			"x-csi-pv-namespace": []string{"my-namespace"},
		}

		req := types.CreateVolumesRequest{
			Volumes: []types.VolumeRequestParam{
				{
					CreateNew: &types.CreateVolumeParam{
						CreateNewFromAttributes: &types.CreateNewFromAttributes{
							VolumeSize:   1024,
							CapacityUnit: "GB",
						},
					},
				},
			},
		}

		_, err = c.CreateVolume(context.Background(), systemID, req, authHeader)
		assert.NoError(t, err)
	})

	t.Run("No headers when opts is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify auth headers are NOT present
			assert.Empty(t, r.Header.Get("x-csi-pv-claimname"), "Should not have PVC claim name header")
			assert.Empty(t, r.Header.Get("x-csi-pv-name"), "Should not have PV name header")
			assert.Empty(t, r.Header.Get("x-csi-pv-namespace"), "Should not have PVC namespace header")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(successBody))
		}))
		defer server.Close()

		c, err := NewClientWithArgs(server.URL, "", true, true, "")
		assert.NoError(t, err)
		c.SetAllowedArrays([]string{allowedArray})

		req := types.CreateVolumesRequest{
			Volumes: []types.VolumeRequestParam{
				{
					CreateNew: &types.CreateVolumeParam{
						CreateNewFromAttributes: &types.CreateNewFromAttributes{
							VolumeSize:   1024,
							CapacityUnit: "GB",
						},
					},
				},
			},
		}

		// Call without opts
		_, err = c.CreateVolume(context.Background(), systemID, req)
		assert.NoError(t, err)
	})

	t.Run("Empty header map does not cause errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(successBody))
		}))
		defer server.Close()

		c, err := NewClientWithArgs(server.URL, "", true, true, "")
		assert.NoError(t, err)
		c.SetAllowedArrays([]string{allowedArray})

		req := types.CreateVolumesRequest{
			Volumes: []types.VolumeRequestParam{
				{
					CreateNew: &types.CreateVolumeParam{
						CreateNewFromAttributes: &types.CreateNewFromAttributes{
							VolumeSize:   1024,
							CapacityUnit: "GB",
						},
					},
				},
			},
		}

		// Pass empty header
		_, err = c.CreateVolume(context.Background(), systemID, req, http.Header{})
		assert.NoError(t, err)
	})
}

func TestPublishMaskingViews(t *testing.T) {
	allowedArray := "testSymID"
	systemID := "testSymID"

	tests := []struct {
		name        string
		systemID    string
		param       *types.PublishMaskingViewsParam
		httpStatus  int
		httpBody    string
		expectedErr string
		wantErr     bool
	}{
		{
			name:     "Successful publish",
			systemID: systemID,
			param: &types.PublishMaskingViewsParam{
				MaskingViews: []types.MaskingViewPublishParam{
					{
						ID: "CSI-Test-MV",
						StorageGroup: &types.StorageGroupPublishParam{
							ID: "CSI-Test-SG",
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"http_status_code": 200,
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": [
						{
							"status": "success",
							"request_id": "CSI-Test-MV",
							"resource_id": "MaskingView"
						}
					]
				}
			}`,
			wantErr: false,
		},
		{
			name:     "None succeeded",
			systemID: systemID,
			param: &types.PublishMaskingViewsParam{
				MaskingViews: []types.MaskingViewPublishParam{
					{ID: "CSI-Test-MV"},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"summary": {
					"total": 1,
					"succeeded": 0,
					"failed": 1,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": []
				}
			}`,
			expectedErr: "none succeeded",
			wantErr:     true,
		},
		{
			name:     "HTTP error response",
			systemID: systemID,
			param: &types.PublishMaskingViewsParam{
				MaskingViews: []types.MaskingViewPublishParam{
					{ID: "CSI-Test-MV"},
				},
			},
			httpStatus:  http.StatusInternalServerError,
			httpBody:    `{}`,
			expectedErr: "500",
			wantErr:     true,
		},
		{
			name:     "Invalid array",
			systemID: "ignoredArray",
			param: &types.PublishMaskingViewsParam{
				MaskingViews: []types.MaskingViewPublishParam{
					{ID: "CSI-Test-MV"},
				},
			},
			expectedErr: "is ignored as it is not managed",
			wantErr:     true,
		},
		{
			name:     "Publish with host and port group",
			systemID: systemID,
			param: &types.PublishMaskingViewsParam{
				MaskingViews: []types.MaskingViewPublishParam{
					{
						ID: "CSI-Test-MV-Full",
						StorageGroup: &types.StorageGroupPublishParam{
							ID:  "CSI-Test-SG",
							SRP: "SRP_1",
						},
						Host: &types.HostPublishParam{
							ID: "CSI-Test-Host",
							Actions: &types.HostPublishActions{
								AddInitiatorsToHostAction: &types.AddInitiatorsToHostAction{
									Initiators: []types.InitiatorPublishParam{
										{ID: "iqn.1993-08.org.debian:01:test"},
									},
								},
							},
						},
						PortGroup: &types.PortGroupPublishParam{
							ID:       "CSI-Test-PG",
							Protocol: "SCSI_FC",
							Actions: &types.PortGroupPublishActions{
								AddPortsToPortGroupAction: &types.AddPortsToPortGroupAction{
									Ports: []types.PortPublishParam{
										{ID: "FA-1D:4"},
									},
								},
							},
						},
					},
				},
			},
			httpStatus: http.StatusOK,
			httpBody: `{
				"http_status_code": 200,
				"summary": {
					"total": 1,
					"succeeded": 1,
					"failed": 0,
					"rejected": 0,
					"partially_succeeded": 0
				},
				"results": {
					"result": [
						{
							"status": "success",
							"request_id": "CSI-Test-MV-Full",
							"resource_id": "MaskingView"
						}
					]
				}
			}`,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method, "Expected POST method")
				assert.Contains(t, r.URL.Path, "/systems/"+tc.systemID+"/masking-views", "Expected URL path")
				w.WriteHeader(tc.httpStatus)
				w.Write([]byte(tc.httpBody))
			}))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			resp, err := c.PublishMaskingViews(context.Background(), tc.systemID, tc.param)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != "" {
					assert.ErrorContains(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Greater(t, resp.Summary.Succeeded, 0)
			}
		})
	}
}

func TestCreateVolumesErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		resp     *types.CreateVolumesResponse
		expected string
	}{
		{
			name:     "nil response",
			resp:     nil,
			expected: "create volumes failed",
		},
		{
			name: "no results",
			resp: &types.CreateVolumesResponse{
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{},
				},
			},
			expected: "create volumes failed",
		},
		{
			name: "result with no messages",
			resp: &types.CreateVolumesResponse{
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{
						{Status: "failed"},
					},
				},
			},
			expected: "create volumes failed",
		},
		{
			name: "result with code and message",
			resp: &types.CreateVolumesResponse{
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{
						{
							Messages: &types.ResponseMessages{
								Message: []types.ResponseMessage{
									{Code: "ERR001", Message: "Volume creation failed"},
								},
							},
						},
					},
				},
			},
			expected: "ERR001: Volume creation failed",
		},
		{
			name: "result with message only (no code)",
			resp: &types.CreateVolumesResponse{
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{
						{
							Messages: &types.ResponseMessages{
								Message: []types.ResponseMessage{
									{Code: "", Message: "Insufficient SRP capacity"},
								},
							},
						},
					},
				},
			},
			expected: "Insufficient SRP capacity",
		},
		{
			name: "result with empty messages slice",
			resp: &types.CreateVolumesResponse{
				Results: types.CreateVolumesResults{
					Result: []types.CreateVolumeResponseItem{
						{
							Messages: &types.ResponseMessages{
								Message: []types.ResponseMessage{},
							},
						},
					},
				},
			},
			expected: "create volumes failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := createVolumesErrorMessage(tc.resp)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestGetVolumesIdentifiersInStorageGroup(t *testing.T) {
	allowedArray := "testSymID"
	tests := []struct {
		name            string
		symID           string
		storageGroupID  string
		expectedVolumes *types.Volumev1
		expectedStatus  int
		expectedErr     string
		responseBody    string
	}{
		{
			name:           "Valid symID and storage group",
			symID:          "testSymID",
			storageGroupID: "SG1",
			expectedVolumes: &types.Volumev1{
				Volumes: []types.VolumeEnhanced{
					{ID: "00120", Identifier: "csi-ABC-tw-5373a1ebea-reptest"},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "Valid symID with no volumes",
			symID:           "testSymID",
			storageGroupID:  "SG1",
			expectedVolumes: &types.Volumev1{Volumes: []types.VolumeEnhanced{}},
			expectedStatus:  http.StatusOK,
		},
		{
			name:           "Error from IsAllowedArray",
			symID:          "",
			storageGroupID: "SG1",
			expectedErr:    "the requested array () is ignored as it is not managed",
		},
		{
			name:           "Empty storage group ID",
			symID:          "testSymID",
			storageGroupID: "",
			expectedErr:    "storageGroupID is empty",
		},
		{
			name:           "HTTP error from API",
			symID:          "testSymID",
			storageGroupID: "SG1",
			expectedStatus: http.StatusInternalServerError,
			expectedErr:    "Internal Server Error",
		},
		{
			name:           "Invalid JSON response",
			symID:          "testSymID",
			storageGroupID: "SG1",
			expectedStatus: http.StatusOK,
			responseBody:   "not-valid-json",
			expectedErr:    "invalid character",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/univmax/rest/v1/systems/"+tc.symID+"/volumes", r.URL.Path, "Expected URL path")
				if tc.storageGroupID != "" {
					assert.Contains(t, r.URL.RawQuery, "filter=storage_groups.id%20EQ%20"+tc.storageGroupID, "Expected filter in query")
				}
				assert.Contains(t, r.URL.RawQuery, "limit=1000", "Expected limit param")
				assert.Contains(t, r.URL.RawQuery, "expiration_delay_secs=30", "Expected expiration param")
				w.WriteHeader(tc.expectedStatus)
				if tc.responseBody != "" {
					w.Write([]byte(tc.responseBody))
				} else if tc.expectedStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tc.expectedVolumes)
				}
			}))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			volumes, err := c.GetVolumesIdentifiersInStorageGroup(context.Background(), tc.symID, tc.storageGroupID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, volumes)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVolumes, volumes)
			}
		})
	}
}

func TestGetVolumesIdentifiersInStorageGroupPagination(t *testing.T) {
	allowedArray := "testSymID"
	storageGroupID := "SG1"
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		assert.Equal(t, "/univmax/rest/v1/systems/testSymID/volumes", r.URL.Path)
		switch requestCount {
		case 1:
			assert.Equal(t, "", r.URL.Query().Get("resume_token"), "Expected no resume token on first request")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&types.Volumev1{
				Volumes:      []types.VolumeEnhanced{{ID: "001", Identifier: "vol-1"}},
				VolumePaging: types.VolumePaging{RemainingInstances: 1, ResumeToken: "token1"},
			})
		case 2:
			assert.Equal(t, "token1", r.URL.Query().Get("resume_token"), "Expected resume token on second request")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&types.Volumev1{
				Volumes:      []types.VolumeEnhanced{{ID: "002", Identifier: "vol-2"}},
				VolumePaging: types.VolumePaging{RemainingInstances: 1, ResumeToken: "token2"},
			})
		case 3:
			assert.Equal(t, "token2", r.URL.Query().Get("resume_token"), "Expected resume token on third request")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&types.Volumev1{
				Volumes: []types.VolumeEnhanced{{ID: "003", Identifier: "vol-3"}},
			})
		}
	}))
	defer server.Close()

	c, err := NewClientWithArgs(server.URL, "", true, true, "")
	assert.NoError(t, err)
	c.SetAllowedArrays([]string{allowedArray})

	volumes, err := c.GetVolumesIdentifiersInStorageGroup(context.Background(), allowedArray, storageGroupID)
	assert.NoError(t, err)
	assert.Equal(t, 3, requestCount, "Expected 3 paginated requests")
	assert.Equal(t, 3, len(volumes.Volumes), "Expected 3 volumes across all pages")
}

func TestGetHostByInitiator(t *testing.T) {
	allowedArray := "testSymID"

	tests := []struct {
		name         string
		symID        string
		wwpn         string
		setupHandler func(w http.ResponseWriter, r *http.Request)
		expectHost   bool
		expectErr    bool
		expectErrMsg string
	}{
		{
			name:  "Host found by single WWPN",
			symID: allowedArray,
			wwpn:  "5000000000000001",
			setupHandler: func(w http.ResponseWriter, r *http.Request) {
				// Mock GetInitiatorList for the WWPN
				if strings.Contains(r.URL.Path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=5000000000000001") {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.InitiatorList{
						InitiatorIDs: []string{"FA-1D:5000000000000001"},
					})
					return
				}
				if r.URL.Path == "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/initiator/FA-1D:5000000000000001" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Initiator{
						InitiatorID: "FA-1D:5000000000000001",
						Host:        "BFS_Host_Node01",
						HostID:      "BFS_Host_Node01",
					})
					return
				}
				if r.URL.Path == "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/host/BFS_Host_Node01" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Host{
						HostID:     "BFS_Host_Node01",
						Initiators: []string{"5000000000000001"},
					})
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			expectHost:   true,
			expectErr:    false,
			expectErrMsg: "",
		},
		{
			name:  "No host found for WWPN — confirmed absent (nil, nil)",
			symID: allowedArray,
			wwpn:  "5000000000000099",
			setupHandler: func(w http.ResponseWriter, r *http.Request) {
				// Mock GetInitiatorList for the WWPN - returns empty list
				if strings.Contains(r.URL.Path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=5000000000000099") {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.InitiatorList{
						InitiatorIDs: []string{},
					})
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			expectHost: false,
			expectErr:  false,
		},
		{
			name:  "API error during host lookup — returns (nil, error) (RACE-3)",
			symID: allowedArray,
			wwpn:  "5000000000000001",
			setupHandler: func(w http.ResponseWriter, r *http.Request) {
				// Mock GetInitiatorList failure
				if strings.Contains(r.URL.Path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=5000000000000001") {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			expectHost:   false,
			expectErr:    true,
			expectErrMsg: "Internal Server Error",
		},
		{
			name:         "Invalid symID — IsAllowedArray error",
			symID:        "invalidArray",
			wwpn:         "5000000000000001",
			setupHandler: func(_ http.ResponseWriter, _ *http.Request) {},
			expectHost:   false,
			expectErr:    true,
			expectErrMsg: "is not managed",
		},
		{
			name:  "GetInitiatorList API error — returns error",
			symID: allowedArray,
			wwpn:  "5000000000000001",
			setupHandler: func(w http.ResponseWriter, r *http.Request) {
				// GetInitiatorList fails
				if strings.Contains(r.URL.Path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=5000000000000001") {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			expectHost:   false,
			expectErr:    true,
			expectErrMsg: "Internal Server Error",
		},
		{
			name:  "WWPN known but not part of any host — confirmed absent (nil, nil)",
			symID: allowedArray,
			wwpn:  "5000000000000003",
			setupHandler: func(w http.ResponseWriter, r *http.Request) {
				// Mock GetInitiatorList
				if strings.Contains(r.URL.Path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=5000000000000003") {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.InitiatorList{
						InitiatorIDs: []string{"FA-1D:5000000000000003"},
					})
					return
				}
				// Initiator exists but has no Host field
				if r.URL.Path == "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/initiator/FA-1D:5000000000000003" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Initiator{
						InitiatorID: "FA-1D:5000000000000003",
						Host:        "",
					})
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			expectHost: false,
			expectErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.setupHandler))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			host, err := c.GetHostByInitiator(context.Background(), tc.symID, tc.wwpn)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, host)
				if tc.expectErrMsg != "" {
					assert.ErrorContains(t, err, tc.expectErrMsg)
				}
			} else if tc.expectHost {
				assert.NoError(t, err)
				require.NotNil(t, host)
				assert.Equal(t, "BFS_Host_Node01", host.HostID)
			} else {
				// (nil, nil) case - confirmed no host
				assert.NoError(t, err)
				assert.Nil(t, host)
			}
		})
	}
}

func TestGetHostByInitiators(t *testing.T) {
	allowedArray := "testSymID"

	// helper: build a handler that maps initiator IDs to hosts
	makeHandler := func(initiatorToHost map[string]string, hostObj *types.Host) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// Mock GetInitiatorList for any WWPN
			if strings.Contains(path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=") {
				wwpn := r.URL.Query().Get("initiator_hba")
				// Find which initiator ID contains this WWPN
				var foundID string
				for initID := range initiatorToHost {
					if strings.Contains(initID, wwpn) {
						foundID = initID
						break
					}
				}
				if foundID != "" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.InitiatorList{
						InitiatorIDs: []string{foundID},
					})
				} else {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.InitiatorList{
						InitiatorIDs: []string{},
					})
				}
				return
			}
			// Match initiator requests
			for initID, hostName := range initiatorToHost {
				if path == "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/initiator/"+initID {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Initiator{
						InitiatorID: initID,
						Host:        hostName,
						HostID:      hostName,
					})
					return
				}
			}
			// Match host request
			if hostObj != nil && path == "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/host/"+hostObj.HostID {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(hostObj)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}
	}

	tests := []struct {
		name         string
		symID        string
		wwpns        []string
		setupHandler func(w http.ResponseWriter, r *http.Request)
		expectHost   bool
		expectErr    bool
		expectErrMsg string
	}{
		{
			name:  "All WWPNs resolve to same host",
			symID: allowedArray,
			wwpns: []string{"5000000000000001", "5000000000000002"},
			setupHandler: makeHandler(
				map[string]string{
					"FA-1D:5000000000000001": "BFS_Host_Node01",
					"FA-1D:5000000000000002": "BFS_Host_Node01",
				},
				&types.Host{
					HostID:     "BFS_Host_Node01",
					Initiators: []string{"FA-1D:5000000000000001", "FA-1D:5000000000000002"},
				},
			),
			expectHost: true,
			expectErr:  false,
		},
		{
			name:  "WWPNs resolve to different hosts — conflict returns (nil, error)",
			symID: allowedArray,
			wwpns: []string{"5000000000000001", "5000000000000002"},
			setupHandler: func(w http.ResponseWriter, r *http.Request) {
				path := r.URL.Path
				// Mock GetInitiatorList
				if strings.Contains(path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=") {
					wwpn := r.URL.Query().Get("initiator_hba")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.InitiatorList{
						InitiatorIDs: []string{"FA-1D:" + wwpn},
					})
					return
				}
				switch path {
				case "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/initiator/FA-1D:5000000000000001":
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Initiator{InitiatorID: "FA-1D:5000000000000001", Host: "BFS_Host_Node01", HostID: "BFS_Host_Node01"})
				case "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/initiator/FA-1D:5000000000000002":
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Initiator{InitiatorID: "FA-1D:5000000000000002", Host: "BFS_Host_Node02", HostID: "BFS_Host_Node02"})
				case "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/host/BFS_Host_Node01":
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Host{HostID: "BFS_Host_Node01", Initiators: []string{"FA-1D:5000000000000001"}})
				case "/univmax/restapi/100/sloprovisioning/symmetrix/testSymID/host/BFS_Host_Node02":
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(&types.Host{HostID: "BFS_Host_Node02", Initiators: []string{"FA-1D:5000000000000002"}})
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			},
			expectHost:   false,
			expectErr:    true,
			expectErrMsg: "conflict",
		},
		{
			name:  "Empty WWPN list returns (nil, error)",
			symID: allowedArray,
			wwpns: []string{},
			setupHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectHost:   false,
			expectErr:    true,
			expectErrMsg: "no initiators provided",
		},
		{
			name:  "API error on first WWPN — returns (nil, error) (RACE-3)",
			symID: allowedArray,
			wwpns: []string{"5000000000000001"},
			setupHandler: func(w http.ResponseWriter, r *http.Request) {
				// Mock GetInitiatorList failure
				if strings.Contains(r.URL.Path, "/initiator") && strings.Contains(r.URL.RawQuery, "initiator_hba=") {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			expectHost: false,
			expectErr:  true,
		},
		{
			name:  "All WWPNs have no host — confirmed absent (nil, nil)",
			symID: allowedArray,
			wwpns: []string{"5000000000000001", "5000000000000002"},
			setupHandler: makeHandler(
				map[string]string{
					"FA-1D:5000000000000001": "",
					"FA-1D:5000000000000002": "",
				},
				nil,
			),
			expectHost: false,
			expectErr:  false,
		},
		{
			name:         "Invalid symID — returns error",
			symID:        "invalidArray",
			wwpns:        []string{"5000000000000001"},
			setupHandler: func(_ http.ResponseWriter, _ *http.Request) {},
			expectHost:   false,
			expectErr:    true,
			expectErrMsg: "is not managed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.setupHandler))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			host, err := c.GetHostByInitiators(context.Background(), tc.symID, tc.wwpns)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, host)
				if tc.expectErrMsg != "" {
					assert.ErrorContains(t, err, tc.expectErrMsg)
				}
			} else if tc.expectHost {
				assert.NoError(t, err)
				require.NotNil(t, host)
				assert.Equal(t, "BFS_Host_Node01", host.HostID)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, host)
			}
		})
	}
}

func TestGetHostMaskingViews(t *testing.T) {
	allowedArray := "testSymID"

	tests := []struct {
		name         string
		symID        string
		hostID       string
		setupHandler func(w http.ResponseWriter, r *http.Request)
		expectedMVs  []string
		expectErr    bool
		expectErrMsg string
	}{
		{
			name:   "Host with masking views",
			symID:  allowedArray,
			hostID: "BFS_Host_Node01",
			setupHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&types.Host{
					HostID:         "BFS_Host_Node01",
					MaskingviewIDs: []string{"MV_Boot_Node01", "MV_CSI_Node01"},
				})
			},
			expectedMVs: []string{"MV_Boot_Node01", "MV_CSI_Node01"},
			expectErr:   false,
		},
		{
			name:   "Host with no masking views",
			symID:  allowedArray,
			hostID: "BFS_Host_Standalone",
			setupHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&types.Host{
					HostID:         "BFS_Host_Standalone",
					MaskingviewIDs: nil,
				})
			},
			expectedMVs: nil,
			expectErr:   false,
		},
		{
			name:   "API error fetching host",
			symID:  allowedArray,
			hostID: "BFS_Host_Node01",
			setupHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedMVs:  nil,
			expectErr:    true,
			expectErrMsg: "Internal Server Error",
		},
		{
			name:         "Invalid symID",
			symID:        "invalidArray",
			hostID:       "BFS_Host_Node01",
			setupHandler: func(_ http.ResponseWriter, _ *http.Request) {},
			expectedMVs:  nil,
			expectErr:    true,
			expectErrMsg: "is not managed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.setupHandler))
			defer server.Close()

			c, err := NewClientWithArgs(server.URL, "", true, true, "")
			assert.NoError(t, err)
			c.SetAllowedArrays([]string{allowedArray})

			mvIDs, err := c.GetHostMaskingViews(context.Background(), tc.symID, tc.hostID)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, mvIDs)
				if tc.expectErrMsg != "" {
					assert.ErrorContains(t, err, tc.expectErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMVs, mvIDs)
			}
		})
	}
}

// TestHostConflictError verifies the typed conflict error reports every host involved
// with its matching WWPNs, and identifies itself as a logical conflict so callers can
// reject rather than retry (FR-1.2).
func TestHostConflictError(t *testing.T) {
	err := &HostConflictError{
		SymID: "000197900111",
		Hosts: map[string][]string{
			"BFS_Host_Node02": {"5000000000000002"},
			"BFS_Host_Node01": {"5000000000000003", "5000000000000001"},
		},
	}

	assert.True(t, err.IsHostConflict())

	msg := err.Error()
	assert.Contains(t, msg, "conflict")
	assert.Contains(t, msg, "000197900111")
	assert.Contains(t, msg, "2 different hosts")
	// Hosts and their WWPNs are sorted so the message is stable across runs.
	assert.Contains(t, msg, "host BFS_Host_Node01 has WWPNs [5000000000000001 5000000000000003]")
	assert.Contains(t, msg, "host BFS_Host_Node02 has WWPNs [5000000000000002]")
	assert.Less(t, strings.Index(msg, "BFS_Host_Node01"), strings.Index(msg, "BFS_Host_Node02"))

	// Callers detect the condition through errors.As even when wrapped.
	var conflict *HostConflictError
	assert.True(t, errors.As(fmt.Errorf("adoption failed: %w", err), &conflict))
	assert.Equal(t, "000197900111", conflict.SymID)
}
