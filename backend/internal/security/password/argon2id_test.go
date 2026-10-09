package password

import "testing"

func TestArgon2idHashAndVerify(t *testing.T) {
	hasher := NewArgon2id()
	first, err := hasher.Hash("CorrectHorse1!")
	if err != nil {
		t.Fatal(err)
	}
	second, err := hasher.Hash("CorrectHorse1!")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("password hashes must use unique salts")
	}

	valid, err := hasher.Verify("CorrectHorse1!", first)
	if err != nil || !valid {
		t.Fatalf("expected matching password, valid=%v err=%v", valid, err)
	}
	valid, err = hasher.Verify("WrongPassword1!", first)
	if err != nil || valid {
		t.Fatalf("expected mismatched password, valid=%v err=%v", valid, err)
	}
}

func TestArgon2idRejectsMalformedHash(t *testing.T) {
	valid, err := NewArgon2id().Verify("Password1!", "not-a-hash")
	if err == nil || valid {
		t.Fatalf("expected malformed hash error, valid=%v err=%v", valid, err)
	}
}
