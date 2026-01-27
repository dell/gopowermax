/*
 *
 * Copyright © 2021-2025 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	types "github.com/dell/gopowermax/v2/types/v100"
	"github.com/stretchr/testify/assert"
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
