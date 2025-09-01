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

/*
Copyright © 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v100

import "testing"

func TestGetJobResource(t *testing.T) {
	tests := []struct {
		name                 string
		job                  Job
		expectedSymmetrixID  string
		expectedResourceType string
		expectedResourceID   string
	}{
		{
			name: "valid resource link",
			job: Job{
				ResourceLink: "provisioning/system/SYMMETRIX-1234/volume/1234",
			},
			expectedSymmetrixID:  "SYMMETRIX-1234",
			expectedResourceType: "volume",
			expectedResourceID:   "1234",
		},
		{
			name: "Empty resource link",
			job: Job{
				ResourceLink: "",
			},
			expectedSymmetrixID:  "",
			expectedResourceType: "",
			expectedResourceID:   "",
		},
		{
			name: "invalid resource link",
			job: Job{
				ResourceLink: "system/SYMMETRIX-1234",
			},
			expectedSymmetrixID:  "",
			expectedResourceType: "",
			expectedResourceID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symID, resourceType, id := tt.job.GetJobResource()

			if symID != tt.expectedSymmetrixID {
				t.Errorf("expected %s, got %s", tt.expectedSymmetrixID, symID)
			}

			if resourceType != tt.expectedResourceType {
				t.Errorf("expected %s, got %s", tt.expectedResourceType, resourceType)
			}

			if id != tt.expectedResourceID {
				t.Errorf("expected %s, got %s", tt.expectedResourceID, id)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
	}{
		{
			name: "valid error",
			err: &Error{
				Message: "test-error",
			},
		},
		{
			name: "empty error",
			err: &Error{
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.err.Message {
				t.Errorf("expected %s, got %s", tt.err.Message, tt.err.Error())
			}
		})
	}
}
