package main

import "testing"

func TestShouldQuarantine(t *testing.T) {
	cases := []struct {
		name  string
		event paymentEvent
		want  bool
	}{
		{"retries exhausted", paymentEvent{Attempts: 3, Amount: 100}, true},
		{"high value", paymentEvent{Attempts: 1, Amount: 100000}, true},
		{"ordinary retry", paymentEvent{Attempts: 2, Amount: 5000}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldQuarantine(tc.event); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
