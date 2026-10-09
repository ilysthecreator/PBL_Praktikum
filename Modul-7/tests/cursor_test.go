package tests

import (
	"testing"
	"time"

	"api-students/pagination"
)

func TestCursorEncodeDecode(t *testing.T) {
	now := time.Now().UTC()
	id := 42

	encoded := pagination.EncodeCursor(now, id)
	if encoded == "" {
		t.Fatal("encoded cursor tidak boleh kosong")
	}

	decoded, err := pagination.DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeCursor gagal: %v", err)
	}

	if decoded.ID != id {
		t.Errorf("ID cursor tidak cocok: got %d, want %d", decoded.ID, id)
	}

	// UnixNano precision match
	if decoded.CreatedAt.UnixNano() != now.UnixNano() {
		t.Errorf("CreatedAt tidak cocok: got %v, want %v", decoded.CreatedAt, now)
	}
}

func TestCursorDecodeInvalid(t *testing.T) {
	invalidCases := []string{
		"",
		"   ",
		"bukanbase64!!",
		"YWJj",       // base64 dari "abc" tanpa pipe
		"MTIzfGFiYw", // base64 dari "123|abc" ID bukan angka
	}

	for _, tc := range invalidCases {
		_, err := pagination.DecodeCursor(tc)
		if err == nil {
			t.Errorf("DecodeCursor(%q) seharusnya gagal", tc)
		}
	}
}
