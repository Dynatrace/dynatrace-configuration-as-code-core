// @license
// Copyright 2025 Dynatrace LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buckets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
)

type StatusClient interface {
	Get(context.Context, string) (api.Response, error)
}

const stateActive = "active"

// AwaitBucketStable waits until the bucket is stable, meaning it's not creating, updating or deleting, but active.
//
// aborts/returns when
//   - the maxDuration is reached.
//   - a client error occurred.
//   - the bucket is stable.
//
// returns
//   - bucketExists: if the bucket exists after the stable check.
//   - err: any possible occurring error.
func AwaitBucketStable(ctx context.Context, client StatusClient, bucketName string, maxDuration time.Duration, durationBetweenTries time.Duration) (bucketExists bool, err error) {
	logger := logr.FromContextOrDiscard(ctx)
	ctx, cancel := context.WithTimeout(ctx, maxDuration)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("context canceled before bucket '%s' became stable", bucketName)
		default:
			// query bucket
			apiResp, err := client.Get(ctx, bucketName)

			if err != nil {
				if api.IsNotFoundError(err) {
					// bucket deleted.
					return false, nil
				}
				apiErr := api.APIError{}
				if !errors.Is(err, &apiErr) {
					return false, err
				}
			} else {
				// try to unmarshal into internal struct
				res, err := unmarshalJSON(apiResp.Data)
				if err != nil {
					return false, err
				}

				if res.Status == stateActive {
					return true, nil
				}
			}

			logger.V(1).Info(fmt.Sprintf("Waiting for bucket '%s' to become stable...", bucketName))
			time.Sleep(durationBetweenTries)
		}
	}
}
