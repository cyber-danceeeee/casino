package notificationchannel

import "testing"

func TestNormalize(t *testing.T) {
	t.Parallel()
	if got, want := Normalize(" fcm,IN_APP,fcm "), "FCM,IN_APP"; got != want {
		t.Fatalf("Normalize() = %q, want %q", got, want)
	}
}

func TestValid(t *testing.T) {
	t.Parallel()
	if !Valid("IN_APP,FCM") {
		t.Fatal("expected combined supported channels to be valid")
	}
	if Valid("FCM,SMS") {
		t.Fatal("expected unsupported channel to be invalid")
	}
}
