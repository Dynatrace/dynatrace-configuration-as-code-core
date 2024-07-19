package api_test

import (
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/stretchr/testify/assert"
)

func TestMultipartResponse_IsSuccess(t *testing.T) {
	t.Run("invoke over null object", func(t *testing.T) {
		var mr *api.MultipartResponse

		assert.False(t, mr.IsSuccess())
	})
}
