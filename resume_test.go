package discordgo

import (
	"sync/atomic"
	"testing"
)

// TestSetResumeState verifies that SetResumeState injects sessionID, sequence,
// and resumeGatewayURL into a Session without opening a connection. No network
// calls are made.
func TestSetResumeState(t *testing.T) {
	cases := []struct {
		name             string
		sessionID        string
		sequence         int64
		resumeGatewayURL string
		wantSessionID    string
		wantSequence     int64
		wantGateway      string
		priorGateway     string
	}{
		{
			name:             "full state injection",
			sessionID:        "abc123",
			sequence:         9001,
			resumeGatewayURL: "wss://gateway-us-east1.discord.gg/?v=10",
			wantSessionID:    "abc123",
			wantSequence:     9001,
			wantGateway:      "wss://gateway-us-east1.discord.gg/?v=10",
		},
		{
			name:             "empty resumeGatewayURL preserves prior gateway",
			sessionID:        "def456",
			sequence:         42,
			resumeGatewayURL: "",
			priorGateway:     "wss://gateway.discord.gg/?v=10",
			wantSessionID:    "def456",
			wantSequence:     42,
			wantGateway:      "wss://gateway.discord.gg/?v=10",
		},
		{
			name:             "zero sequence is injected",
			sessionID:        "ghi789",
			sequence:         0,
			resumeGatewayURL: "wss://us-east1-c.gateway.discord.gg",
			wantSessionID:    "ghi789",
			wantSequence:     0,
			wantGateway:      "wss://us-east1-c.gateway.discord.gg",
		},
		{
			name:             "large sequence",
			sessionID:        "jkl000",
			sequence:         1<<62 - 1,
			resumeGatewayURL: "wss://gateway-001.discord.gg",
			wantSessionID:    "jkl000",
			wantSequence:     1<<62 - 1,
			wantGateway:      "wss://gateway-001.discord.gg",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Session{sequence: new(int64)}
			if tc.priorGateway != "" {
				s.gateway = tc.priorGateway
			}

			s.SetResumeState(tc.sessionID, tc.sequence, tc.resumeGatewayURL)

			if s.sessionID != tc.wantSessionID {
				t.Errorf("sessionID = %q, want %q", s.sessionID, tc.wantSessionID)
			}
			gotSeq := atomic.LoadInt64(s.sequence)
			if gotSeq != tc.wantSequence {
				t.Errorf("sequence = %d, want %d", gotSeq, tc.wantSequence)
			}
			if s.gateway != tc.wantGateway {
				t.Errorf("gateway = %q, want %q", s.gateway, tc.wantGateway)
			}
		})
	}
}

// TestSetResumeState_OpenCondition verifies that after SetResumeState with a
// non-empty sessionID and non-zero sequence the resume branch would be taken
// by Open() (condition: sessionID != "" || sequence != 0). No network call.
func TestSetResumeState_OpenCondition(t *testing.T) {
	s := &Session{sequence: new(int64)}
	s.SetResumeState("sess-xyz", 100, "")

	wouldResume := !(s.sessionID == "" && atomic.LoadInt64(s.sequence) == 0)
	if !wouldResume {
		t.Error("Open() would IDENTIFY but expected RESUME after SetResumeState")
	}
}
