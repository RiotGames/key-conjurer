package command

import (
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"
)

func Test_tryParseTimeToLiveError(t *testing.T) {
	t.Run("UnambiguousAmount", func(t *testing.T) {
		validationError := smithy.GenericAPIError{
			Code:    "ValidationError",
			Message: "1 validation error detected: Value '86400' at 'durationSeconds' failed to satisfy constraint: Member must have value less than or equal to 43200",
		}
		err, ok := tryParseTimeToLiveError(&validationError)

		require.True(t, ok)
		require.NotNil(t, err)
		require.Equal(t, err.Error(), "you requested a TTL of 24 hours, but the maximum for this configuration is 12 hours")
		var ttlError TimeToLiveError
		require.ErrorAs(t, err, &ttlError)
		require.Equal(t, ttlError.MaxDuration, 43200*time.Second)
		require.Equal(t, ttlError.RequestedDuration, 86400*time.Second)
		require.Equal(t, ttlError.Code(), ExitCodeValueError)
	})

	t.Run("AmbiguousAmount", func(t *testing.T) {
		validationError := smithy.GenericAPIError{
			Code:    "ValidationError",
			Message: "The requested DurationSeconds exceeds the MaxSessionDuration set for this role.",
		}
		err, ok := tryParseTimeToLiveError(&validationError)

		require.True(t, ok)
		require.NotNil(t, err)
		require.Equal(t, err.Error(), "the TTL you requested exceeds the maximum TTL for this configuration")
		var ttlError TimeToLiveError
		require.ErrorAs(t, err, &ttlError)
		require.Equal(t, ttlError.MaxDuration, time.Duration(0))
		require.Equal(t, ttlError.RequestedDuration, time.Duration(0))
		require.Equal(t, ttlError.Code(), ExitCodeValueError)
	})
}
