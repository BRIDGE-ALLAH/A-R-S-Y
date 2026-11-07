package httpapi

// OnboardingStatus contains only non-sensitive first-run setup state.
type OnboardingStatus struct {
	SuperuserConfigured bool `json:"superuser_configured"`
	SecretKeyConfigured bool `json:"secret_key_configured"`
	ServerRegistered   bool `json:"server_registered"`
}

func onboardingStatusFromCounts(superuserCount, serverCount int64, secretConfigured bool) OnboardingStatus {
	return OnboardingStatus{
		SuperuserConfigured: superuserCount > 0,
		SecretKeyConfigured: secretConfigured,
		ServerRegistered: serverCount > 0,
	}
}
