package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_GetAllCounters(t *testing.T) {
	tests := []struct {
		name     string
		counters map[string]int64
		want     map[string]int64
	}{
		{
			name: "non-empty counters",
			counters: map[string]int64{
				"counter1": 10,
				"counter2": 20,
			},
			want: map[string]int64{
				"counter1": 10,
				"counter2": 20,
			},
		},
		{
			name:     "empty counters",
			counters: map[string]int64{},
			want:     map[string]int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				counters: tt.counters,
			}
			assert.Equal(t, tt.want, s.GetAllCounters())
		})
	}
}

func TestMemStorage_GetAllGauges(t *testing.T) {
	tests := []struct {
		name   string
		gauges map[string]float64
		want   map[string]float64
	}{
		{
			name: "non-empty gauges",
			gauges: map[string]float64{
				"gauge1": 1.23,
				"gauge2": 4.56,
			},
			want: map[string]float64{
				"gauge1": 1.23,
				"gauge2": 4.56,
			},
		},
		{
			name:   "empty gauges",
			gauges: map[string]float64{},
			want:   map[string]float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				gauges: tt.gauges,
			}
			assert.Equal(t, tt.want, s.GetAllGauges())
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	tests := []struct {
		name     string
		counters map[string]int64
		arg      string
		want     int64
		wantOk   bool
	}{
		{
			name: "existing counter",
			counters: map[string]int64{
				"counter1": 10,
			},
			arg:    "counter1",
			want:   10,
			wantOk: true,
		},
		{
			name:     "non-existing counter",
			counters: map[string]int64{},
			arg:      "nonexistent",
			want:     0,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				counters: tt.counters,
			}
			got, ok := s.GetCounter(tt.arg)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	tests := []struct {
		name   string
		gauges map[string]float64
		arg    string
		want   float64
		wantOk bool
	}{
		{
			name: "existing gauge",
			gauges: map[string]float64{
				"gauge1": 1.23,
			},
			arg:    "gauge1",
			want:   1.23,
			wantOk: true,
		},
		{
			name:   "non-existing gauge",
			gauges: map[string]float64{},
			arg:    "nonexistent",
			want:   0.0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				gauges: tt.gauges,
			}
			got, ok := s.GetGauge(tt.arg)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestMemStorage_Reset(t *testing.T) {
	s := &MemStorage{
		gauges: map[string]float64{
			"gauge1": 1.23,
		},
		counters: map[string]int64{
			"counter1": 10,
		},
	}

	s.Reset()

	assert.Empty(t, s.gauges)
	assert.Empty(t, s.counters)
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]int64
		argName  string
		argValue int64
		want     map[string]int64
	}{
		{
			name: "update_counter_success",
			initial: map[string]int64{
				"counter1": 10,
			},
			argName:  "counter1",
			argValue: 20,
			want: map[string]int64{
				"counter1": 30,
			},
		},
		{
			name:     "add_counter_success",
			initial:  map[string]int64{},
			argName:  "counter2",
			argValue: 30,
			want: map[string]int64{
				"counter2": 30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				counters: tt.initial,
			}
			s.UpdateCounter(tt.argName, tt.argValue)
			assert.Equal(t, tt.want, s.counters)
		})
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]float64
		argName  string
		argValue float64
		want     map[string]float64
	}{
		{
			name: "update_gauge_success",
			initial: map[string]float64{
				"gauge1": 1.23,
			},
			argName:  "gauge1",
			argValue: 2.34,
			want: map[string]float64{
				"gauge1": 2.34,
			},
		},
		{
			name:     "add_gauge_success",
			initial:  map[string]float64{},
			argName:  "gauge2",
			argValue: 4.56,
			want: map[string]float64{
				"gauge2": 4.56,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				gauges: tt.initial,
			}
			s.UpdateGauge(tt.argName, tt.argValue)
			assert.Equal(t, tt.want, s.gauges)
		})
	}
}

func TestNewMemStorage(t *testing.T) {
	t.Run("create empty storage", func(t *testing.T) {
		got := NewMemStorage()
		assert.NotNil(t, got)
		assert.Empty(t, got.gauges)
		assert.Empty(t, got.counters)
	})
}
