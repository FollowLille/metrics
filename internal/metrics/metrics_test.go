package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRuntimeMetrics(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "get_runtime_metrics",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricsFull := GetRuntimeMetrics()
			assert.NotEmpty(t, metricsFull, "GetRuntimeMetrics() is empty")
		},
		)
	}
}
