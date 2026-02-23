package biz

import "testing"

func TestCreditAccount_ReserveSettleReleaseFlow(t *testing.T) {
	account := NewCreditAccount(1001)
	account.AddCredits(200)

	if err := account.ReserveCredits(120); err != nil {
		t.Fatalf("reserve credits failed: %v", err)
	}
	if account.Balance != 80 {
		t.Fatalf("unexpected balance after reserve: got %d want %d", account.Balance, 80)
	}
	if account.ReservedBalance != 120 {
		t.Fatalf("unexpected reserved balance after reserve: got %d want %d", account.ReservedBalance, 120)
	}

	if err := account.SettleReservedCredits(50); err != nil {
		t.Fatalf("settle reserved credits failed: %v", err)
	}
	if account.Balance != 80 {
		t.Fatalf("unexpected balance after settle: got %d want %d", account.Balance, 80)
	}
	if account.ReservedBalance != 70 {
		t.Fatalf("unexpected reserved balance after settle: got %d want %d", account.ReservedBalance, 70)
	}
	if account.TotalConsumed != 50 {
		t.Fatalf("unexpected consumed total after settle: got %d want %d", account.TotalConsumed, 50)
	}

	if err := account.ReleaseReservedCredits(70); err != nil {
		t.Fatalf("release reserved credits failed: %v", err)
	}
	if account.Balance != 150 {
		t.Fatalf("unexpected balance after release: got %d want %d", account.Balance, 150)
	}
	if account.ReservedBalance != 0 {
		t.Fatalf("unexpected reserved balance after release: got %d want %d", account.ReservedBalance, 0)
	}
}

func TestCreditAccount_ReserveSettleReleaseValidation(t *testing.T) {
	account := NewCreditAccount(1002)
	account.AddCredits(20)

	if err := account.ReserveCredits(0); err != ErrInvalidCreditAmount {
		t.Fatalf("reserve with zero amount should fail: got %v", err)
	}

	if err := account.ReserveCredits(30); err != ErrInsufficientBalance {
		t.Fatalf("reserve with insufficient balance should fail: got %v", err)
	}

	if err := account.ReserveCredits(10); err != nil {
		t.Fatalf("reserve should succeed: %v", err)
	}

	if err := account.SettleReservedCredits(20); err != ErrInsufficientReserved {
		t.Fatalf("settle with insufficient reserved should fail: got %v", err)
	}

	if err := account.ReleaseReservedCredits(20); err != ErrInsufficientReserved {
		t.Fatalf("release with insufficient reserved should fail: got %v", err)
	}
}
