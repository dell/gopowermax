/*
 Copyright © 2026 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package inttest

import (
	"context"
	"sync"
	"testing"

	"github.com/dell/gopowermax/v2/api"
	"github.com/stretchr/testify/assert"
)

// integrationTestObserver captures RequestObservations from gopowermax during integration tests.
type integrationTestObserver struct {
	mu           sync.Mutex
	observations []api.RequestObservation
}

func (o *integrationTestObserver) ObservePowerMaxRequest(obs api.RequestObservation) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.observations = append(o.observations, obs)
}

func (o *integrationTestObserver) getObservations() []api.RequestObservation {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]api.RequestObservation(nil), o.observations...)
}

func TestRequestObserverWithGetSymmetrixByID(t *testing.T) {
	if client == nil {
		t.Skip("PowerMax client not initialized")
	}

	observer := &integrationTestObserver{}
	client.SetRequestObserver(observer)
	defer client.SetRequestObserver(nil)

	_, err := client.GetSymmetrixByID(context.Background(), symmetrixID)
	if err != nil {
		t.Skipf("Unable to verify observer against live array: %v", err)
	}

	observations := observer.getObservations()
	assert.NotEmpty(t, observations)

	last := observations[len(observations)-1]
	assert.Equal(t, "GET", last.Method)
	assert.Equal(t, 200, last.StatusCode)
	assert.NotEmpty(t, last.Endpoint)
	assert.Nil(t, last.Err)
	assert.Greater(t, last.Duration.Nanoseconds(), int64(0))
}
