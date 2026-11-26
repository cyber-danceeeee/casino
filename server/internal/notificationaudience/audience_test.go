package notificationaudience

import (
	"reflect"
	"testing"
)

func TestValues(t *testing.T) {
	t.Parallel()
	want := []string{"3", "8", "13"}
	got := Values(" 3,8\n13, ")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Values() = %#v, want %#v", got, want)
	}
}
