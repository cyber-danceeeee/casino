package notificationpush

import "testing"

func TestPlainText(t *testing.T) {
	t.Parallel()
	if got, want := plainText("<p>Hello&nbsp; <b>player</b></p>"), "Hello player"; got != want {
		t.Fatalf("plainText() = %q, want %q", got, want)
	}
}

func TestTruncateText(t *testing.T) {
	t.Parallel()
	if got, want := truncateText("通知消息", 2), "通知"; got != want {
		t.Fatalf("truncateText() = %q, want %q", got, want)
	}
}
