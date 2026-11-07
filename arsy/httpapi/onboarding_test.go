package httpapi

import "testing"

func TestOnboardingStatusFromCounts(t *testing.T) {
	tests := []struct {
		name string
		superusers int64
		servers int64
		secretConfigured bool
		want OnboardingStatus
	}{
		{
			name: "fresh install",
			want: OnboardingStatus{},
		},
		{
			name: "configured control plane",
			superusers: 1,
			servers: 2,
			secretConfigured: true,
			want: OnboardingStatus{SuperuserConfigured: true, SecretKeyConfigured: true, ServerRegistered: true},
		},
		{
			name: "secret only",
			secretConfigured: true,
			want: OnboardingStatus{SecretKeyConfigured: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := onboardingStatusFromCounts(tt.superusers, tt.servers, tt.secretConfigured)
			if got != tt.want {
				t.Fatalf("status = %#v, want %#v", got, tt.want)
			}
		})
	}
}
