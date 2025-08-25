/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
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
