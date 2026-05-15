package api

import (
	"testing"

	"k8s.io/client-go/rest"
)

func TestApplyRateLimitOverrides(t *testing.T) {
	tests := []struct {
		name      string
		qpsEnv    string
		burstEnv  string
		wantQPS   float32
		wantBurst int
	}{
		{
			name:      "defaults preserved when env unset",
			qpsEnv:    "",
			burstEnv:  "",
			wantQPS:   0,
			wantBurst: 0,
		},
		{
			name:      "valid values applied",
			qpsEnv:    "50",
			burstEnv:  "100",
			wantQPS:   50,
			wantBurst: 100,
		},
		{
			name:      "fractional qps applied",
			qpsEnv:    "12.5",
			burstEnv:  "25",
			wantQPS:   12.5,
			wantBurst: 25,
		},
		{
			name:      "invalid qps ignored, burst applied",
			qpsEnv:    "notanumber",
			burstEnv:  "20",
			wantQPS:   0,
			wantBurst: 20,
		},
		{
			name:      "non-positive values ignored",
			qpsEnv:    "0",
			burstEnv:  "-1",
			wantQPS:   0,
			wantBurst: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GLANCE_KUBE_API_QPS", tc.qpsEnv)
			t.Setenv("GLANCE_KUBE_API_BURST", tc.burstEnv)

			cfg := &rest.Config{}
			applyRateLimitOverrides(cfg)

			if cfg.QPS != tc.wantQPS {
				t.Errorf("QPS: got %v, want %v", cfg.QPS, tc.wantQPS)
			}
			if cfg.Burst != tc.wantBurst {
				t.Errorf("Burst: got %d, want %d", cfg.Burst, tc.wantBurst)
			}
		})
	}
}
