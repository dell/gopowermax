/*
 Copyright © 2021 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package api

import (
	"net/http"
	"reflect"
	"testing"
)

type stubTypeWithMetaData struct{}

func (s stubTypeWithMetaData) MetaData() http.Header {
	h := make(http.Header)
	h.Set("foo", "bar")
	return h
}

func Test_addMetaData(t *testing.T) {
	tests := []struct {
		name           string
		givenHeader    map[string]string
		expectedHeader map[string]string
		body           interface{}
	}{
		{"nil header is a noop", nil, nil, nil},
		{"nil body is a noop", nil, nil, nil},
		{"header is updated", make(map[string]string), map[string]string{"Foo": "bar"}, stubTypeWithMetaData{}},
		{"header is not updated", make(map[string]string), map[string]string{}, struct{}{}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			addMetaData(tt.givenHeader, tt.body)

			switch {
			case tt.givenHeader == nil:
				if tt.givenHeader != nil {
					t.Errorf("(%s): expected %s, actual %s", tt.body, tt.expectedHeader, tt.givenHeader)
				}
			case tt.body == nil:
				if len(tt.givenHeader) != 0 {
					t.Errorf("(%s): expected %s, actual %s", tt.body, tt.expectedHeader, tt.givenHeader)
				}
			default:
				if !reflect.DeepEqual(tt.expectedHeader, tt.givenHeader) {
					t.Errorf("(%s): expected %s, actual %s", tt.body, tt.expectedHeader, tt.givenHeader)
				}
			}
		})
	}
}
