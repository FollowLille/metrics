package agent

import (
	"github.com/FollowLille/metrics/internal/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgent_ChangeAddress(t *testing.T) {
	type fields struct {
		ServerAddress      string
		ServerPort         int64
		PollCount          int64
		PollInterval       time.Duration
		ReportSendInterval time.Duration
		metrics            map[string]float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    string
		want    string
		wantErr bool
	}{
		{
			name: "change_address",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    "http://127.0.0.1",
			want:    "127.0.0.1",
			wantErr: false,
		},
		{
			name: "change_address_without_prefix",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    "http://example.com",
			want:    "example.com",
			wantErr: false,
		},
		{
			name: "change_address_with_port_error",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    "http://127.0.0.1:8090",
			want:    "localhost",
			wantErr: true,
		},
		{
			name: "error_change_address",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    "localhost:8080",
			want:    "localhost",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				ServerAddress:      tt.fields.ServerAddress,
				ServerPort:         tt.fields.ServerPort,
				PollCount:          tt.fields.PollCount,
				PollInterval:       tt.fields.PollInterval,
				ReportSendInterval: tt.fields.ReportSendInterval,
				metrics:            tt.fields.metrics,
			}

			err := a.ChangeAddress(tt.args)
			if tt.wantErr {
				assert.Error(t, err, "Agent.ChangeAddress() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				assert.Equal(t, tt.want, a.ServerAddress, "Agent.ChangeAddress() name = %v, current = %v, want %v", tt.name, a.ServerAddress, tt.want)
			} else {
				assert.NoError(t, err, "Agent.ChangeAddress() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				assert.Equal(t, tt.want, a.ServerAddress, "Agent.ChangeAddress() name = %v, current = %v, want %v", tt.name, a.ServerAddress, tt.want)
			}
		})
	}
}

func TestAgent_ChangeIntervalByName(t *testing.T) {
	type fields struct {
		ServerAddress      string
		ServerPort         int64
		PollCount          int64
		PollInterval       time.Duration
		ReportSendInterval time.Duration
		metrics            map[string]float64
	}
	type args struct {
		name    string
		seconds int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "change_interval",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    args{name: "poll", seconds: 5},
			want:    5 * time.Second,
			wantErr: false,
		},
		{
			name: "error_change_poll_interval",
			fields: fields{
				ServerAddress:      "localhost",
				ServerPort:         8080,
				PollCount:          0,
				PollInterval:       2 * time.Second,
				ReportSendInterval: 10 * time.Second,
				metrics:            make(map[string]float64),
			},
			args:    args{name: "poll", seconds: 0},
			want:    2 * time.Second,
			wantErr: true,
		},
		{
			name: "error_change_report_interval",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    args{name: "report", seconds: 0},
			want:    10 * time.Second,
			wantErr: true,
		},
		{
			name: "error_change_non_exists_interval",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    args{name: "random_interval", seconds: 0},
			want:    10 * time.Second,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				ServerAddress:      tt.fields.ServerAddress,
				ServerPort:         tt.fields.ServerPort,
				PollCount:          tt.fields.PollCount,
				PollInterval:       tt.fields.PollInterval,
				ReportSendInterval: tt.fields.ReportSendInterval,
				metrics:            tt.fields.metrics,
			}
			err := a.ChangeIntervalByName(tt.args.name, tt.args.seconds)
			if wantErr := tt.wantErr; wantErr {
				assert.Error(t, err, "Agent.ChangeIntervalByName() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
			} else {
				assert.NoError(t, err, "Agent.ChangeIntervalByName() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				if tt.args.name == "poll" {
					assert.Equal(t, tt.want, a.PollInterval, "Agent.ChangeIntervalByName() name = %v, error = %v, want %v", tt.name, a.PollInterval, tt.want)
				} else if tt.args.name == "report" {
					assert.Equal(t, tt.want, a.ReportSendInterval, "Agent.ChangeIntervalByName() name = %v, error = %v, want %v", tt.name, a.ReportSendInterval, tt.want)
				}
			}
		})
	}
}

func TestAgent_ChangePort(t *testing.T) {
	type fields struct {
		ServerAddress      string
		ServerPort         int64
		PollCount          int64
		PollInterval       time.Duration
		ReportSendInterval time.Duration
		metrics            map[string]float64
	}
	type args struct {
		port int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "change_port",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    args{port: 8081},
			want:    8081,
			wantErr: false,
		},
		{
			name: "error_change_port_less_1024",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			args:    args{port: 0},
			want:    8080,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				ServerAddress:      tt.fields.ServerAddress,
				ServerPort:         tt.fields.ServerPort,
				PollCount:          tt.fields.PollCount,
				PollInterval:       tt.fields.PollInterval,
				ReportSendInterval: tt.fields.ReportSendInterval,
				metrics:            tt.fields.metrics,
			}
			err := a.ChangePort(tt.args.port)
			if tt.wantErr {
				assert.Error(t, a.ChangePort(tt.args.port), "Agent.ChangePort() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				assert.Equal(t, tt.want, a.ServerPort, "Agent.ChangePort() name = %v, port = %v, want %v", tt.name, a.ServerPort, tt.want)
			} else {
				assert.NoError(t, err, "Agent.ChangePort() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				assert.Equal(t, tt.want, a.ServerPort, "Agent.ChangePort() name = %v, port = %v, want %v", tt.name, a.ServerPort, tt.want)
			}
		})
	}
}

func TestAgent_GetMetrics(t *testing.T) {
	type fields struct {
		ServerAddress      string
		ServerPort         int64
		PollCount          int64
		PollInterval       time.Duration
		ReportSendInterval time.Duration
		metrics            map[string]float64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "not_empty_metrics",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				ServerAddress:      tt.fields.ServerAddress,
				ServerPort:         tt.fields.ServerPort,
				PollCount:          tt.fields.PollCount,
				PollInterval:       tt.fields.PollInterval,
				ReportSendInterval: tt.fields.ReportSendInterval,
				metrics:            tt.fields.metrics,
			}
			a.GetMetrics()
			assert.NotEmpty(t, a.metrics, "Agent.GetMetrics() is empty")
		})
	}
}

func TestAgent_SendMetrics(t *testing.T) {
	type fields struct {
		ServerAddress      string
		ServerPort         int64
		PollCount          int64
		PollInterval       time.Duration
		ReportSendInterval time.Duration
		metrics            map[string]float64
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "send_metrics",
			fields: fields{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				ServerAddress:      tt.fields.ServerAddress,
				ServerPort:         tt.fields.ServerPort,
				PollCount:          tt.fields.PollCount,
				PollInterval:       tt.fields.PollInterval,
				ReportSendInterval: tt.fields.ReportSendInterval,
				metrics:            tt.fields.metrics,
			}
			err := a.SendMetrics()
			if tt.wantErr {
				assert.Error(t, a.SendMetrics(), "Agent.SendMetrics() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
			} else {
				assert.NoError(t, err, "Agent.SendMetrics() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
func TestNewAgent(t *testing.T) {
	tests := []struct {
		name string
		want *Agent
	}{
		{
			name: "equal_new_agent",
			want: &Agent{
				ServerAddress:      config.Address,
				ServerPort:         config.Port,
				PollCount:          0,
				PollInterval:       config.PollInterval,
				ReportSendInterval: config.ReportSendInterval,
				metrics:            make(map[string]float64)},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, NewAgent(), "NewAgent() name = %v, wantErr %v", tt.name, tt.want)
	}
}
