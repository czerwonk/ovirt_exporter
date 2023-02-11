// SPDX-License-Identifier: MIT

package collector

import "context"

type Client interface {
	// GetAndParse retrieves XML data from the API and unmarshals it
	GetAndParse(ctx context.Context, path string, v interface{}) error
}
