package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FollowLille/metrics/internal/config"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name string
		want *Server
	}{
		{
			name: "new_server",
			want: &Server{
				Address: config.Address,
				Port:    config.Port,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			assert.Equal(t, tt.want, server)
		})
	}
}

func TestServer_ChangeAddress(t *testing.T) {
	type fields struct {
		Address string
		Port    int64
	}
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change_address_success",
			fields: fields{
				Address: config.Address,
				Port:    config.Port,
			},
			args:    args{address: "http://125.0.0.1"},
			wantErr: false,
		},
		{
			name: "change_address_without_prefix_error",
			fields: fields{
				Address: config.Address,
				Port:    config.Port,
			},
			args:    args{address: "example.com"},
			wantErr: true,
		},
		{
			name: "change_address_error",
			fields: fields{
				Address: config.Address,
				Port:    config.Port,
			},
			args:    args{address: ":8080"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Server{
				Address: tt.fields.Address,
				Port:    tt.fields.Port,
			}
			err := c.ChangeAddress(tt.args.address)
			if tt.wantErr {
				assert.Error(t, err)
				assert.NotEqual(t, tt.args.address, c.Address, "ChangeAddress() name = %v, current = %v, want %v", tt.name, c.Address, tt.args.address)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.args.address, c.Address, "ChangeAddress() name = %v, current = %v, want %v", tt.name, c.Address, tt.args.address)
			}
		})
	}
}

func TestServer_ChangePort(t *testing.T) {
	type fields struct {
		Address string
		Port    int64
	}
	type args struct {
		port int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change_port_success",
			fields: fields{
				Address: config.Address,
				Port:    config.Port,
			},
			args:    args{port: 8080},
			wantErr: false,
		},
		{
			name: "change_port_error",
			fields: fields{
				Address: config.Address,
				Port:    config.Port,
			},
			args:    args{port: 65536},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Server{
				Address: tt.fields.Address,
				Port:    tt.fields.Port,
			}
			err := c.ChangePort(tt.args.port)
			if tt.wantErr {
				assert.Error(t, err)
				assert.NotEqual(t, tt.args.port, c.Port, "ChangePort() name = %v, current = %v, want %v", tt.name, c.Port, tt.args.port)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.args.port, c.Port, "ChangePort() name = %v, current = %v, want %v", tt.name, c.Port, tt.args.port)
			}
		})
	}
}
