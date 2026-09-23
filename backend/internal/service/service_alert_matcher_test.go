package service

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

func TestEffectiveTriggerPrice(t *testing.T) {
	target := 90.0
	cases := []struct {
		name    string
		alert   model.PriceAlert
		want    float64
		hasWant bool
	}{
		{
			name:    "no condition",
			alert:   model.PriceAlert{BaselinePrice: 100},
			hasWant: false,
		},
		{
			name:    "target only",
			alert:   model.PriceAlert{TargetPrice: target, BaselinePrice: 100},
			want:    90,
			hasWant: true,
		},
		{
			name:    "drop only",
			alert:   model.PriceAlert{DropPercent: 10, BaselinePrice: 100},
			want:    90,
			hasWant: true,
		},
		{
			name:    "both, drop price is lower",
			alert:   model.PriceAlert{TargetPrice: 95, DropPercent: 10, BaselinePrice: 100},
			want:    90,
			hasWant: true,
		},
		{
			name:    "both, target is lower",
			alert:   model.PriceAlert{TargetPrice: 80, DropPercent: 5, BaselinePrice: 100},
			want:    80,
			hasWant: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := effectiveTriggerPrice(tc.alert)
			if !tc.hasWant {
				if got != nil {
					t.Fatalf("expected nil trigger price, got %v", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected trigger price %v, got nil", tc.want)
			}
			if *got != tc.want {
				t.Fatalf("expected trigger price %v, got %v", tc.want, *got)
			}
		})
	}
}

func TestMatchAlert(t *testing.T) {
	alert := model.PriceAlert{TargetPrice: 90, DropPercent: 10, BaselinePrice: 100, Status: constants.AlertStatusActive}
	// 有效触发价为 max(90, 90)=90
	if matched, price := matchAlert(alert, 90); !matched || price != 90 {
		t.Fatalf("expected match at 90, got matched=%v price=%v", matched, price)
	}
	if matched, _ := matchAlert(alert, 91); matched {
		t.Fatalf("price above trigger must not match")
	}
	if matched, _ := matchAlert(alert, 0); matched {
		t.Fatalf("no in-stock offer must not match")
	}
}
