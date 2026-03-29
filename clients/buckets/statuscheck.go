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
	"log/slog"
	"time"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
)

// StatusClient is the interface required by AwaitActiveOrNotFound to poll a bucket's status.
type StatusClient interface {
	Get(context.Context, string) (api.Response, error)
}

const stateActive = "active"

// AwaitActiveOrNotFound waits until the bucket is active or deleted, meaning it's not creating, updating or deleting.
//
// It returns when:
//   - the maxDuration is reached
//   - a non-API client error occurs
//   - the bucket is active or has been removed (404)
//
// Returns:
//   - bucketExists: whether the bucket exists after the check
//   - err: any error that occurred
func AwaitActiveOrNotFound(ctx context.Context, client StatusClient, bucketName string, maxDuration time.Duration, durationBetweenTries time.Duration) (bucketExists bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, maxDuration)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return false, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "context canceled before bucket became stable"}
		default:
			// query bucket
			apiResp, err := client.Get(ctx, bucketName)

			if err != nil {
				if api.IsNotFoundError(err) {
					// bucket deleted.
					return false, nil
				}
				var apiErr api.APIError
				if !errors.As(err, &apiErr) {
					return false, err
				}
				sleep(ctx, bucketName, durationBetweenTries)
				continue
			}
			// try to unmarshal into internal struct
			res, err := unmarshalJSON(apiResp.Data)
			if err != nil {
				return false, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to unmarshal bucket response", Wrapped: err}
			}

			if res.Status == stateActive {
				return true, nil
			}
			sleep(ctx, bucketName, durationBetweenTries)
		}
	}
}

func sleep(ctx context.Context, bucketName string, durationBetweenTries time.Duration) {
	slog.DebugContext(ctx, "Waiting for bucket to become stable", slog.String("bucketName", bucketName))
	time.Sleep(durationBetweenTries)
}
