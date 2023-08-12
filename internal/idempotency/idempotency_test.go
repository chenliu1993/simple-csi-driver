package idempotency

import (
	"sync"
	"testing"
)

func initTestIdentity() *Idempotency {
	return &Idempotency{
		lock: &sync.Mutex{},
		processing: map[string]bool{
			"testVolumeId1": true,
		},
	}

}

func TestIsProcessing(t *testing.T) {
	idempotency := initTestIdentity()
	tcs := []struct {
		description string
		volumeId    string
		expected    bool
	}{
		{
			description: "testVolumeId1 expects to be processing",
			volumeId:    "testVolumeId1",
			expected:    true,
		},
		{
			description: "testVolumeId2 expects not to be processing",
			volumeId:    "testVolumeId2",
			expected:    false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			if idempotency.IsProcessing(tc.volumeId) != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, idempotency.IsProcessing(tc.volumeId))
			}
		})
	}
}

func TestAddProcessing(t *testing.T) {
	idempotency := initTestIdentity()
	idempotency.AddProcessing("testVolumeId2")

	actualResult := idempotency.IsProcessing("testVolumeId2")
	if actualResult != true {
		t.Errorf("Expected %v, got %v", true, actualResult)
	}
}

func TestRemoveProcessing(t *testing.T) {
	idempotency := initTestIdentity()
	idempotency.RemoveProcessing("testVolumeId1")

	actualResult := idempotency.IsProcessing("testVolumeId1")
	if actualResult != false {
		t.Errorf("Expected %v, got %v", false, actualResult)
	}
}
