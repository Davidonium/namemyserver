package namemyserver_test

import (
	"fmt"
	"testing"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

func TestValidateNameTable(t *testing.T) {
	cases := []struct {
		Input string
		Valid bool
	}{
		{
			Input: "test",
			Valid: true,
		},
		{
			Input: "teset-with-dashes",
			Valid: true,
		},
		{
			Input: "test-with-dashes-at-the-end--",
			Valid: false,
		},
		{
			Input: "---test-with-dashes-at-the-start",
			Valid: false,
		},
		{
			Input: "UPPERCASE",
			Valid: false,
		},
		{
			Input: "Test-with-mixed-Case",
			Valid: false,
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("Test Case #%d", i), func(t *testing.T) {
			ok := namemyserver.ValidateName(tt.Input)
			if ok != tt.Valid {
				t.Errorf("ValidateName() = input: %q - got %v, want %v", tt.Input, ok, tt.Valid)
			}
		})
	}
}

func TestValidateNameSegmentTable(t *testing.T) {
	cases := []struct {
		Input string
		Valid bool
	}{
		{
			Input: "test",
			Valid: true,
		},
		{
			Input: "00test",
			Valid: true,
		},
		{
			Input: "test01",
			Valid: true,
		},
		{
			Input: "test01test",
			Valid: true,
		},
		{
			Input: "test-with-dashes",
			Valid: false,
		},
		{
			Input: "UPPERCASE",
			Valid: false,
		},
		{
			Input: "TestWithMixedCase",
			Valid: false,
		},
		{
			Input: "TestWithMixedCaseAndNumbers000",
			Valid: false,
		},
		{
			Input: "000TestWithMixedCaseAndNumbers",
			Valid: false,
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("Test Case #%d", i), func(t *testing.T) {
			ok := namemyserver.ValidateNameSegment(tt.Input)
			if ok != tt.Valid {
				t.Errorf("ValidateNameSegment() = input: %q - got %v, want %v", tt.Input, ok, tt.Valid)
			}
		})
	}
}
