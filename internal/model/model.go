package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Crockford base32 alphabet for ULID encoding.
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// UUIDVersion identifies the UUID variant.
type UUIDVersion int

const (
	VersionUnknown UUIDVersion = 0
	Version1       UUIDVersion = 1
	Version2       UUIDVersion = 2
	Version3       UUIDVersion = 3
	Version4       UUIDVersion = 4
	Version5       UUIDVersion = 5
	Version6       UUIDVersion = 6
	Version7       UUIDVersion = 7
	Version8       UUIDVersion = 8
)

func (v UUIDVersion) String() string {
	switch v {
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	case 5:
		return "5"
	case 6:
		return "6"
	case 7:
		return "7"
	case 8:
		return "8"
	default:
		return "unknown"
	}
}

// Variant describes the UUID variant.
type Variant string

const (
	VariantNCS       Variant = "ncs"
	VariantRFC4122   Variant = "rfc4122"
	VariantMicrosoft Variant = "microsoft"
	VariantReserved  Variant = "reserved"
)

// UUID represents a parsed UUID.
type UUID struct {
	Raw     [16]byte
	Version UUIDVersion
	Variant Variant
}

// ULID represents a parsed ULID.
type ULID struct {
	Raw       [16]byte
	Timestamp int64
}

// --- UUID Generation ---

// GenerateUUIDv4 creates a random (version 4) UUID.
func GenerateUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return FormatUUID(b), nil
}

// GenerateUUIDv7 creates a time-ordered (version 7) UUID.
// The first 48 bits encode the Unix timestamp in milliseconds,
// followed by 74 bits of randomness.
func GenerateUUIDv7() (string, error) {
	var b [16]byte
	now := time.Now().UnixMilli()
	b[0] = byte(now >> 40)
	b[1] = byte(now >> 32)
	b[2] = byte(now >> 24)
	b[3] = byte(now >> 16)
	b[4] = byte(now >> 8)
	b[5] = byte(now)
	if _, err := rand.Read(b[6:]); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x70 // version 7
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return FormatUUID(b), nil
}

// FormatUUID formats 16 raw bytes into the canonical 8-4-4-4-12 hex string.
func FormatUUID(b [16]byte) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}

// --- ULID Generation ---

// GenerateULID creates a 26-character Crockford base32 ULID.
// The first 48 bits encode the Unix timestamp in milliseconds,
// followed by 80 bits of randomness.
func GenerateULID() (string, error) {
	var b [16]byte
	now := time.Now().UnixMilli()
	b[0] = byte(now >> 40)
	b[1] = byte(now >> 32)
	b[2] = byte(now >> 24)
	b[3] = byte(now >> 16)
	b[4] = byte(now >> 8)
	b[5] = byte(now)
	if _, err := rand.Read(b[6:]); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return EncodeCrockford(b), nil
}

// EncodeCrockford encodes 16 bytes into 26 Crockford base32 characters.
func EncodeCrockford(b [16]byte) string {
	var result [26]byte
	bitPos := 0
	for i := 0; i < 26; i++ {
		var val byte
		for j := 0; j < 5; j++ {
			val <<= 1
			if bitPos < 128 {
				byteIdx := bitPos / 8
				bitIdx := 7 - (bitPos % 8)
				val |= (b[byteIdx] >> bitIdx) & 1
			}
			bitPos++
		}
		result[i] = crockford[val]
	}
	return string(result[:])
}

// --- UUID Parsing ---

// ParseUUID parses a canonical UUID string and extracts metadata.
func ParseUUID(s string) (*UUID, error) {
	if len(s) != 36 {
		return nil, fmt.Errorf("invalid UUID length: expected 36 characters (8-4-4-4-12), got %d", len(s))
	}
	// Check dash positions
	for _, pos := range []int{8, 13, 18, 23} {
		if s[pos] != '-' {
			return nil, fmt.Errorf("invalid UUID format: expected dash at position %d", pos)
		}
	}
	hexStr := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID hex: %w", err)
	}
	var raw [16]byte
	copy(raw[:], b)

	version := UUIDVersion(raw[6] >> 4)
	var variant Variant
	switch raw[8] >> 6 {
	case 0, 1:
		variant = VariantNCS
	case 2:
		variant = VariantRFC4122
	case 3:
		if raw[8]&0x20 != 0 {
			variant = VariantReserved
		} else {
			variant = VariantMicrosoft
		}
	}

	return &UUID{Raw: raw, Version: version, Variant: variant}, nil
}

// --- ULID Parsing ---

// ParseULID decodes a 26-character Crockford base32 ULID string.
func ParseULID(s string) (*ULID, error) {
	if len(s) != 26 {
		return nil, fmt.Errorf("invalid ULID length: expected 26 characters, got %d", len(s))
	}
	b, err := DecodeCrockford(s)
	if err != nil {
		return nil, err
	}
	ts := int64(b[0])<<40 | int64(b[1])<<32 | int64(b[2])<<24 |
		int64(b[3])<<16 | int64(b[4])<<8 | int64(b[5])
	return &ULID{Raw: b, Timestamp: ts}, nil
}

// DecodeCrockford decodes a 26-character Crockford base32 string into 16 bytes.
func DecodeCrockford(s string) ([16]byte, error) {
	var result [16]byte
	var lookup [256]byte
	for i := range lookup {
		lookup[i] = 0xFF
	}
	for i := 0; i < len(crockford); i++ {
		lookup[crockford[i]] = byte(i)
		// Accept lowercase
		if crockford[i] >= 'A' && crockford[i] <= 'Z' {
			lookup[crockford[i]+32] = byte(i)
		}
	}
	bitPos := 0
	for _, c := range []byte(s) {
		val := lookup[c]
		if val == 0xFF {
			return result, fmt.Errorf("invalid ULID character: %q", string(c))
		}
		for j := 4; j >= 0; j-- {
			if bitPos < 128 {
				bit := (val >> j) & 1
				byteIdx := bitPos / 8
				bitIdx := 7 - (bitPos % 8)
				result[byteIdx] |= bit << bitIdx
			}
			bitPos++
		}
	}
	return result, nil
}

// --- Validation ---

// ValidateUUID checks whether a string is a valid UUID.
func ValidateUUID(s string) bool {
	_, err := ParseUUID(s)
	return err == nil
}

// ValidateULID checks whether a string is a valid ULID.
func ValidateULID(s string) bool {
	_, err := ParseULID(s)
	return err == nil
}

// --- Detection ---

// DetectType determines whether a string is a UUID or ULID.
func DetectType(s string) string {
	if _, err := ParseUUID(s); err == nil {
		return "uuid"
	}
	if _, err := ParseULID(s); err == nil {
		return "ulid"
	}
	return "unknown"
}
