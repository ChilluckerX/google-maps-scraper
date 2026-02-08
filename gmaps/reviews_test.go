package gmaps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_extractPlaceID_FromPlaceDataURL(t *testing.T) {
	u := "https://www.google.com/maps/place/D'+Tunggal+Seafood/@0,0,15z/data=!4m6!3m5!1s0x31b7bc18e2a60973:0x67987bf1c83699f1!8m2!3d5.3430086!4d103.1018952!16s%2Fg%2F11c1pcnkq1?hl=en&entry=ttu"

	placeID, err := extractPlaceID(u)
	require.NoError(t, err)
	require.Equal(t, "0x31b7bc18e2a60973:0x67987bf1c83699f1", placeID)
}
