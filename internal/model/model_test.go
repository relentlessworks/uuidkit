package model

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var ulidRegex = regexp.MustCompile(`^[0-9A-Z]{26}$`)

func TestGenerateUUIDv4(t *testing.T) {
	id, err := GenerateUUIDv4()
	if err != nil {
		t.Fatalf("GenerateUUIDv4() error: %v", err)
	}
	if !uuidRegex.MatchString(id) {
		t.Errorf("GenerateUUIDv4() = %q, does not match UUID format", id)
	}
	u, err := ParseUUID(id)
	if err != nil {
		t.Fatalf("ParseUUID(%q) error: %v", id, err)
	}
	if u.Version != Version4 {
		t.Errorf("version = %d, want 4", u.Version)
	}
	if u.Variant != VariantRFC4122 {
		t.Errorf("variant = %s, want rfc4122", u.Variant)
	}
}

func TestGenerateUUIDv4Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		id, err := GenerateUUIDv4()
		if err != nil {
			t.Fatalf("GenerateUUIDv4() error: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate UUID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestGenerateUUIDv7(t *testing.T) {
	id, err := GenerateUUIDv7()
	if err != nil {
		t.Fatalf("GenerateUUIDv7() error: %v", err)
	}
	if !uuidRegex.MatchString(id) {
		t.Errorf("GenerateUUIDv7() = %q, does not match UUID format", id)
	}
	u, err := ParseUUID(id)
	if err != nil {
		t.Fatalf("ParseUUID(%q) error: %v", id, err)
	}
	if u.Version != Version7 {
		t.Errorf("version = %d, want 7", u.Version)
	}
	if u.Variant != VariantRFC4122 {
		t.Errorf("variant = %s, want rfc4122", u.Variant)
	}
}

func TestGenerateUUIDv7Timestamp(t *testing.T) {
	before := time.Now().UnixMilli()
	id, err := GenerateUUIDv7()
	if err != nil {
		t.Fatalf("GenerateUUIDv7() error: %v", err)
	}
	after := time.Now().UnixMilli()

	u, err := ParseUUID(id)
	if err != nil {
		t.Fatalf("ParseUUID(%q) error: %v", id, err)
	}
	// Extract 48-bit timestamp from first 6 bytes
	ts := int64(u.Raw[0])<<40 | int64(u.Raw[1])<<32 | int64(u.Raw[2])<<24 |
		int64(u.Raw[3])<<16 | int64(u.Raw[4])<<8 | int64(u.Raw[5])
	if ts < before || ts > after {
		t.Errorf("timestamp %d not in range [%d, %d]", ts, before, after)
	}
}

func TestGenerateUUIDv7Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		id, err := GenerateUUIDv7()
		if err != nil {
			t.Fatalf("GenerateUUIDv7() error: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate UUID v7 generated: %s", id)
		}
		seen[id] = true
	}
}

func TestGenerateULID(t *testing.T) {
	id, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID() error: %v", err)
	}
	if !ulidRegex.MatchString(id) {
		t.Errorf("GenerateULID() = %q, does not match ULID format", id)
	}
}

func TestGenerateULIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		id, err := GenerateULID()
		if err != nil {
			t.Fatalf("GenerateULID() error: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate ULID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestGenerateULIDTimestamp(t *testing.T) {
	before := time.Now().UnixMilli()
	id, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID() error: %v", err)
	}
	after := time.Now().UnixMilli()

	u, err := ParseULID(id)
	if err != nil {
		t.Fatalf("ParseULID(%q) error: %v", id, err)
	}
	if u.Timestamp < before || u.Timestamp > after {
		t.Errorf("timestamp %d not in range [%d, %d]", u.Timestamp, before, after)
	}
}

func TestParseUUID(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		version UUIDVersion
		variant Variant
	}{
		// UUID v4 with RFC 4122 variant (byte 8 = 0xa7, top 2 bits = 10)
		{"550e8400-e29b-41d4-a716-446655440000", false, Version4, VariantRFC4122},
		// UUID v7 with RFC 4122 variant (byte 8 = 0x8e, top 2 bits = 10)
		{"01892b3a-7e3f-7e3f-8e3f-446655440000", false, Version7, VariantRFC4122},
		// UUID v1 with NCS variant (byte 8 = 0x00, top 2 bits = 00)
		{"00000000-0000-1000-0000-000000000000", false, Version1, VariantNCS},
		// UUID v3 with NCS variant (byte 8 = 0x00, top 2 bits = 00)
		{"00000000-0000-3000-0000-000000000000", false, Version3, VariantNCS},
		// UUID v5 with NCS variant (byte 8 = 0x00, top 2 bits = 00)
		{"00000000-0000-5000-0000-000000000000", false, Version5, VariantNCS},
		// UUID v4 with RFC 4122 variant (byte 8 = 0x80, top 2 bits = 10)
		{"00000000-0000-4000-8000-000000000000", false, Version4, VariantRFC4122},
		// Invalid inputs
		{"not-a-uuid", true, VersionUnknown, ""},
		{"550e8400-e29b-41d4-a716", true, VersionUnknown, ""},
		{"550e8400e29b41d4a716446655440000", true, VersionUnknown, ""},
		{"", true, VersionUnknown, ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			u, err := ParseUUID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUUID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil {
				if u.Version != tt.version {
					t.Errorf("version = %d, want %d", u.Version, tt.version)
				}
				if u.Variant != tt.variant {
					t.Errorf("variant = %s, want %s", u.Variant, tt.variant)
				}
			}
		})
	}
}

func TestParseULID(t *testing.T) {
	// Generate a ULID and parse it back
	id, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID() error: %v", err)
	}
	u, err := ParseULID(id)
	if err != nil {
		t.Fatalf("ParseULID(%q) error: %v", id, err)
	}
	if u.Timestamp <= 0 {
		t.Errorf("timestamp = %d, want > 0", u.Timestamp)
	}

	// Test invalid ULIDs
	// Note: Crockford base32 excludes I, L, O, U
	tests := []struct {
		input   string
		wantErr bool
	}{
		// Valid ULID (all chars in Crockford alphabet, 26 chars)
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2S", false},
		// Too short
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2", true},
		// Too long
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2ST", true},
		// 'I' is not in Crockford
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2I", true},
		// 'U' is not in Crockford
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2U", true},
		// 'L' is not in Crockford
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2L", true},
		// 'O' is not in Crockford
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2O", true},
		// Empty
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseULID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseULID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"01892b3a-7e3f-7e3f-8e3f-446655440000", true},
		{"not-a-uuid", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidateUUID(tt.input); got != tt.want {
				t.Errorf("ValidateUUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateULID(t *testing.T) {
	id, _ := GenerateULID()
	tests := []struct {
		input string
		want  bool
	}{
		{id, true},
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2S", true},
		{"not-a-ulid", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ValidateULID(tt.input); got != tt.want {
				t.Errorf("ValidateULID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectType(t *testing.T) {
	uuidID, _ := GenerateUUIDv4()
	ulidID, _ := GenerateULID()
	tests := []struct {
		input string
		want  string
	}{
		{uuidID, "uuid"},
		{ulidID, "ulid"},
		{"550e8400-e29b-41d4-a716-446655440000", "uuid"},
		{"01H8ZG3F4V5J6K7M8N9P0Q1R2S", "ulid"},
		{"not-valid", "unknown"},
		{"", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := DetectType(tt.input); got != tt.want {
				t.Errorf("DetectType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEncodeDecodeCrockfordRoundTrip(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id, err := GenerateULID()
		if err != nil {
			t.Fatalf("GenerateULID() error: %v", err)
		}
		decoded, err := DecodeCrockford(id)
		if err != nil {
			t.Fatalf("DecodeCrockford(%q) error: %v", id, err)
		}
		reencoded := EncodeCrockford(decoded)
		if reencoded != id {
			t.Errorf("round-trip failed: %q -> %q", id, reencoded)
		}
	}
}

func TestFormatUUID(t *testing.T) {
	var b [16]byte
	for i := range b {
		b[i] = byte(i)
	}
	got := FormatUUID(b)
	want := "00010203-0405-0607-0809-0a0b0c0d0e0f"
	if got != want {
		t.Errorf("FormatUUID() = %q, want %q", got, want)
	}
}

func TestParseUUIDCaseInsensitive(t *testing.T) {
	upper := "550E8400-E29B-41D4-A716-446655440000"
	lower := strings.ToLower(upper)
	u1, err1 := ParseUUID(upper)
	u2, err2 := ParseUUID(lower)
	if err1 != nil || err2 != nil {
		t.Fatalf("ParseUUID case insensitive failed: %v, %v", err1, err2)
	}
	if u1.Version != u2.Version {
		t.Errorf("version mismatch: %d vs %d", u1.Version, u2.Version)
	}
}

func TestParseULIDCaseInsensitive(t *testing.T) {
	id, _ := GenerateULID()
	lower := strings.ToLower(id)
	_, err := ParseULID(lower)
	if err != nil {
		t.Errorf("ParseULID(lowercase) error: %v", err)
	}
}
