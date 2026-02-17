package main

import (
	"os"
	"testing"
	"time"
)

// =============================================================================
// Test 1.2.x: getEnvInt function tests
// =============================================================================

func TestGetEnvInt_ReturnsDefaultWhenEnvVarNotSet(t *testing.T) {
	// Ensure the env var is not set
	os.Unsetenv("TEST_HP_VAR")

	result := getEnvInt("TEST_HP_VAR", 42)

	if result != 42 {
		t.Errorf("getEnvInt() = %d; want 42", result)
	}
}

func TestGetEnvInt_ReturnsParsedValueWhenEnvVarIsValidInteger(t *testing.T) {
	os.Setenv("TEST_HP_VAR", "100")
	defer os.Unsetenv("TEST_HP_VAR")

	result := getEnvInt("TEST_HP_VAR", 42)

	if result != 100 {
		t.Errorf("getEnvInt() = %d; want 100", result)
	}
}

func TestGetEnvInt_ReturnsDefaultWhenEnvVarIsInvalid(t *testing.T) {
	os.Setenv("TEST_HP_VAR", "not-a-number")
	defer os.Unsetenv("TEST_HP_VAR")

	result := getEnvInt("TEST_HP_VAR", 42)

	if result != 42 {
		t.Errorf("getEnvInt() = %d; want 42 (default)", result)
	}
}

func TestGetEnvInt_ReturnsDefaultWhenEnvVarIsEmpty(t *testing.T) {
	os.Setenv("TEST_HP_VAR", "")
	defer os.Unsetenv("TEST_HP_VAR")

	result := getEnvInt("TEST_HP_VAR", 42)

	if result != 42 {
		t.Errorf("getEnvInt() = %d; want 42 (default)", result)
	}
}

func TestGetEnvInt_HandlesNegativeValues(t *testing.T) {
	os.Setenv("TEST_HP_VAR", "-50")
	defer os.Unsetenv("TEST_HP_VAR")

	result := getEnvInt("TEST_HP_VAR", 42)

	if result != -50 {
		t.Errorf("getEnvInt() = %d; want -50", result)
	}
}

func TestGetEnvInt_HandlesZero(t *testing.T) {
	os.Setenv("TEST_HP_VAR", "0")
	defer os.Unsetenv("TEST_HP_VAR")

	result := getEnvInt("TEST_HP_VAR", 42)

	if result != 0 {
		t.Errorf("getEnvInt() = %d; want 0", result)
	}
}

// =============================================================================
// Test 1.3.x: getURLForProto function tests
// =============================================================================

func TestGetURLForProto_ReturnsHTTPForProtoHTTP1(t *testing.T) {
	result := getURLForProto("example.com", protoHTTP1)

	expected := "http://example.com"
	if result != expected {
		t.Errorf("getURLForProto() = %q; want %q", result, expected)
	}
}

func TestGetURLForProto_ReturnsHTTPSForProtoHTTPS(t *testing.T) {
	result := getURLForProto("example.com", protoHTTPS)

	expected := "https://example.com"
	if result != expected {
		t.Errorf("getURLForProto() = %q; want %q", result, expected)
	}
}

func TestGetURLForProto_ReturnsHTTPSForProtoHTTP2(t *testing.T) {
	result := getURLForProto("example.com", protoHTTP2)

	expected := "https://example.com"
	if result != expected {
		t.Errorf("getURLForProto() = %q; want %q", result, expected)
	}
}

func TestGetURLForProto_ReturnsHTTPSForProtoHTTP3(t *testing.T) {
	result := getURLForProto("example.com", protoHTTP3)

	expected := "https://example.com"
	if result != expected {
		t.Errorf("getURLForProto() = %q; want %q", result, expected)
	}
}

func TestGetURLForProto_HandlesIPv4Address(t *testing.T) {
	result := getURLForProto("192.168.1.1", protoHTTPS)

	expected := "https://192.168.1.1"
	if result != expected {
		t.Errorf("getURLForProto() = %q; want %q", result, expected)
	}
}

func TestGetURLForProto_HandlesIPv6Address(t *testing.T) {
	result := getURLForProto("[::1]", protoHTTPS)

	expected := "https://[::1]"
	if result != expected {
		t.Errorf("getURLForProto() = %q; want %q", result, expected)
	}
}

func TestGetURLForProto_HandlesHostWithPort(t *testing.T) {
	result := getURLForProto("example.com:8080", protoHTTP1)

	expected := "http://example.com:8080"
	if result != expected {
		t.Errorf("getURLForProto() = %q; want %q", result, expected)
	}
}

// =============================================================================
// Test 1.4.x: getBlock function tests
// =============================================================================

// Helper to save and restore package-level threshold variables
func withThresholds(min, green, yellow int64, fn func()) {
	origMin := minLatency
	origGreen := greenThreshold
	origYellow := yellowThreshold

	minLatency = min
	greenThreshold = green
	yellowThreshold = yellow

	defer func() {
		minLatency = origMin
		greenThreshold = origGreen
		yellowThreshold = origYellow
	}()

	fn()
}

// Helper to extract the block character from the ANSI-colored string
func extractBlock(s string) string {
	// The format is: color + block + reset
	// We need to extract the Unicode block character
	for _, r := range s {
		for _, b := range blocks {
			if string(r) == b {
				return string(r)
			}
		}
	}
	return ""
}

// Helper to check if the result contains the expected color code
func containsColor(s, colorCode string) bool {
	return len(s) >= len(colorCode) && s[:len(colorCode)] == colorCode
}

func TestGetBlock_GreenZone_LowLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency at the low end of green zone
		result := getBlock(10 * time.Millisecond)

		if !containsColor(result, green) {
			t.Errorf("getBlock(10ms) should be green colored")
		}

		block := extractBlock(result)
		if block != "▁" {
			t.Errorf("getBlock(10ms) block = %q; want %q", block, "▁")
		}
	})
}

func TestGetBlock_GreenZone_MidLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency in the middle of green zone
		result := getBlock(75 * time.Millisecond)

		if !containsColor(result, green) {
			t.Errorf("getBlock(75ms) should be green colored")
		}

		block := extractBlock(result)
		// 75ms out of 150ms = 50% progress, should be block index 1 (▂)
		if block != "▂" {
			t.Errorf("getBlock(75ms) block = %q; want %q", block, "▂")
		}
	})
}

func TestGetBlock_GreenZone_HighLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency near the high end of green zone
		result := getBlock(140 * time.Millisecond)

		if !containsColor(result, green) {
			t.Errorf("getBlock(140ms) should be green colored")
		}

		block := extractBlock(result)
		// 140ms out of 150ms = ~93% progress, should be block index 2 (▃)
		if block != "▃" {
			t.Errorf("getBlock(140ms) block = %q; want %q", block, "▃")
		}
	})
}

func TestGetBlock_YellowZone_LowLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency just into yellow zone
		result := getBlock(160 * time.Millisecond)

		if !containsColor(result, yellow) {
			t.Errorf("getBlock(160ms) should be yellow colored")
		}

		block := extractBlock(result)
		// 10ms into 250ms span = ~4% progress, should be block index 3 (▄)
		if block != "▄" {
			t.Errorf("getBlock(160ms) block = %q; want %q", block, "▄")
		}
	})
}

func TestGetBlock_YellowZone_HighLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency near the high end of yellow zone
		result := getBlock(380 * time.Millisecond)

		if !containsColor(result, yellow) {
			t.Errorf("getBlock(380ms) should be yellow colored")
		}

		block := extractBlock(result)
		// 230ms into 250ms span = 92% progress, should be block index 4 (▅)
		if block != "▅" {
			t.Errorf("getBlock(380ms) block = %q; want %q", block, "▅")
		}
	})
}

func TestGetBlock_RedZone_AtThreshold(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency exactly at yellow threshold (start of red zone)
		result := getBlock(400 * time.Millisecond)

		if !containsColor(result, red) {
			t.Errorf("getBlock(400ms) should be red colored")
		}

		block := extractBlock(result)
		// 0ms into red zone, should be block index 5 (▆)
		if block != "▆" {
			t.Errorf("getBlock(400ms) block = %q; want %q", block, "▆")
		}
	})
}

func TestGetBlock_RedZone_HighLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency well into red zone
		result := getBlock(600 * time.Millisecond)

		if !containsColor(result, red) {
			t.Errorf("getBlock(600ms) should be red colored")
		}

		block := extractBlock(result)
		// 200ms into 400ms red span = 50% progress, should be block index 6 (▇)
		if block != "▇" {
			t.Errorf("getBlock(600ms) block = %q; want %q", block, "▇")
		}
	})
}

func TestGetBlock_RedZone_VeryHighLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency far into red zone (should cap at max block)
		result := getBlock(1000 * time.Millisecond)

		if !containsColor(result, red) {
			t.Errorf("getBlock(1000ms) should be red colored")
		}

		block := extractBlock(result)
		// Way beyond 2x yellowThreshold, should cap at block index 7 (█)
		if block != "█" {
			t.Errorf("getBlock(1000ms) block = %q; want %q", block, "█")
		}
	})
}

func TestGetBlock_EdgeCase_ZeroLatency(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		result := getBlock(0)

		if !containsColor(result, green) {
			t.Errorf("getBlock(0) should be green colored")
		}

		block := extractBlock(result)
		// Zero latency should be the smallest block
		if block != "▁" {
			t.Errorf("getBlock(0) block = %q; want %q", block, "▁")
		}
	})
}

func TestGetBlock_EdgeCase_AtMinLatency(t *testing.T) {
	withThresholds(50, 150, 400, func() {
		// Test latency at exactly minLatency
		result := getBlock(50 * time.Millisecond)

		if !containsColor(result, green) {
			t.Errorf("getBlock(50ms) should be green colored")
		}

		block := extractBlock(result)
		// At minLatency, should be the smallest block
		if block != "▁" {
			t.Errorf("getBlock(50ms) block = %q; want %q", block, "▁")
		}
	})
}

func TestGetBlock_EdgeCase_BelowMinLatency(t *testing.T) {
	withThresholds(50, 150, 400, func() {
		// Test latency below minLatency
		result := getBlock(30 * time.Millisecond)

		if !containsColor(result, green) {
			t.Errorf("getBlock(30ms) should be green colored")
		}

		block := extractBlock(result)
		// Below minLatency, should still be the smallest block
		if block != "▁" {
			t.Errorf("getBlock(30ms) block = %q; want %q", block, "▁")
		}
	})
}

func TestGetBlock_EdgeCase_AtGreenThreshold(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency exactly at green threshold (transition point)
		result := getBlock(150 * time.Millisecond)

		// At exactly greenThreshold, it should be yellow (ms >= greenThreshold enters yellow)
		// Actually, the condition is ms < greenThreshold for green, so 150ms should be yellow
		if !containsColor(result, yellow) {
			t.Errorf("getBlock(150ms) should be yellow colored (at threshold)")
		}

		block := extractBlock(result)
		// Just entered yellow zone, should be block index 3 (▄)
		if block != "▄" {
			t.Errorf("getBlock(150ms) block = %q; want %q", block, "▄")
		}
	})
}

func TestGetBlock_EdgeCase_JustBelowGreenThreshold(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency just below green threshold
		result := getBlock(149 * time.Millisecond)

		if !containsColor(result, green) {
			t.Errorf("getBlock(149ms) should be green colored")
		}
	})
}

func TestGetBlock_EdgeCase_JustBelowYellowThreshold(t *testing.T) {
	withThresholds(0, 150, 400, func() {
		// Test latency just below yellow threshold
		result := getBlock(399 * time.Millisecond)

		if !containsColor(result, yellow) {
			t.Errorf("getBlock(399ms) should be yellow colored")
		}
	})
}

func TestGetBlock_CustomThresholds(t *testing.T) {
	withThresholds(10, 100, 200, func() {
		// Test with custom thresholds

		// Low green
		result := getBlock(20 * time.Millisecond)
		if !containsColor(result, green) {
			t.Errorf("getBlock(20ms) with custom thresholds should be green")
		}

		// Yellow
		result = getBlock(150 * time.Millisecond)
		if !containsColor(result, yellow) {
			t.Errorf("getBlock(150ms) with custom thresholds should be yellow")
		}

		// Red
		result = getBlock(250 * time.Millisecond)
		if !containsColor(result, red) {
			t.Errorf("getBlock(250ms) with custom thresholds should be red")
		}
	})
}

// =============================================================================
// Table-driven tests for more comprehensive coverage
// =============================================================================

func TestGetBlock_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		min           int64
		green         int64
		yellow        int64
		latencyMs     int64
		expectedColor string
		expectedBlock string
	}{
		// Default thresholds (0, 150, 400)
		{"default-green-low", 0, 150, 400, 10, green, "▁"},
		{"default-green-mid", 0, 150, 400, 75, green, "▂"},
		{"default-green-high", 0, 150, 400, 140, green, "▃"},
		{"default-yellow-low", 0, 150, 400, 150, yellow, "▄"},
		{"default-yellow-mid", 0, 150, 400, 275, yellow, "▅"},
		{"default-yellow-high", 0, 150, 400, 380, yellow, "▅"},
		{"default-red-low", 0, 150, 400, 400, red, "▆"},
		{"default-red-mid", 0, 150, 400, 600, red, "▇"},
		{"default-red-high", 0, 150, 400, 900, red, "█"},

		// With minLatency set
		{"min50-at-min", 50, 150, 400, 50, green, "▁"},
		{"min50-below-min", 50, 150, 400, 30, green, "▁"},
		{"min50-above-min", 50, 150, 400, 100, green, "▂"},

		// Edge cases with equal thresholds (span = 0)
		{"narrow-green-span", 100, 100, 400, 50, green, "▁"},
		{"narrow-yellow-span", 0, 150, 150, 150, red, "▆"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withThresholds(tc.min, tc.green, tc.yellow, func() {
				result := getBlock(time.Duration(tc.latencyMs) * time.Millisecond)

				if !containsColor(result, tc.expectedColor) {
					t.Errorf("getBlock(%dms) color mismatch: got result %q, expected color %q",
						tc.latencyMs, result, tc.expectedColor)
				}

				block := extractBlock(result)
				if block != tc.expectedBlock {
					t.Errorf("getBlock(%dms) block = %q; want %q",
						tc.latencyMs, block, tc.expectedBlock)
				}
			})
		})
	}
}

func TestGetURLForProto_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		protoLevel int
		expected   string
	}{
		{"http1-simple", "example.com", protoHTTP1, "http://example.com"},
		{"https-simple", "example.com", protoHTTPS, "https://example.com"},
		{"http2-simple", "example.com", protoHTTP2, "https://example.com"},
		{"http3-simple", "example.com", protoHTTP3, "https://example.com"},
		{"http1-with-port", "example.com:8080", protoHTTP1, "http://example.com:8080"},
		{"https-with-port", "example.com:443", protoHTTPS, "https://example.com:443"},
		{"http1-ipv4", "192.168.1.1", protoHTTP1, "http://192.168.1.1"},
		{"https-ipv4", "8.8.8.8", protoHTTPS, "https://8.8.8.8"},
		{"https-ipv6", "[::1]", protoHTTPS, "https://[::1]"},
		{"https-ipv6-full", "[2001:4860:4860::8888]", protoHTTPS, "https://[2001:4860:4860::8888]"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getURLForProto(tc.host, tc.protoLevel)
			if result != tc.expected {
				t.Errorf("getURLForProto(%q, %d) = %q; want %q",
					tc.host, tc.protoLevel, result, tc.expected)
			}
		})
	}
}

// =============================================================================
// Test: truncateToWidth function tests
// =============================================================================

func TestTruncateToWidth_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		wantVis  int  // expected visible length of result
		wantFull bool // true if result should equal input (no truncation)
	}{
		{"plain-fits", "hello", 10, 5, true},
		{"plain-exact", "hello", 5, 5, true},
		{"plain-truncated", "hello world", 5, 5, false},
		{"ansi-fits", "\033[32mhello\033[0m", 10, 5, true},
		{"ansi-exact", "\033[32mhello\033[0m", 5, 5, true},
		{"ansi-truncated", "\033[32mhello world\033[0m", 5, 5, false},
		{"width-zero", "hello", 0, 0, false},
		{"width-one", "hello", 1, 1, false},
		{"empty-string", "", 10, 0, true},
		{"multi-ansi", "\033[31m\033[1m!\033[0m", 5, 1, true},
		{"multi-ansi-trunc", "\033[31ma\033[0m\033[32mb\033[0m\033[33mc\033[0m", 2, 2, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := truncateToWidth(tc.input, tc.width)

			if tc.wantFull {
				if result != tc.input {
					t.Errorf("truncateToWidth(%q, %d) = %q; want %q (unchanged)",
						tc.input, tc.width, result, tc.input)
				}
				return
			}

			// Count visible characters (non-ANSI)
			vis := 0
			i := 0
			for i < len(result) {
				if result[i] == '\033' && i+1 < len(result) && result[i+1] == '[' {
					j := i + 2
					for j < len(result) && (result[j] < 'A' || result[j] > 'Z') && (result[j] < 'a' || result[j] > 'z') {
						j++
					}
					if j < len(result) {
						j++
					}
					i = j
					continue
				}
				vis++
				i++
			}

			if vis > tc.wantVis {
				t.Errorf("truncateToWidth(%q, %d) visible chars = %d; want <= %d",
					tc.input, tc.width, vis, tc.wantVis)
			}
		})
	}
}

func TestGetEnvInt_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		setEnv   bool
		def      int64
		expected int64
	}{
		{"not-set", "", false, 42, 42},
		{"empty-string", "", true, 42, 42},
		{"valid-positive", "100", true, 42, 100},
		{"valid-zero", "0", true, 42, 0},
		{"valid-negative", "-50", true, 42, -50},
		{"invalid-text", "abc", true, 42, 42},
		{"invalid-float", "3.14", true, 42, 42},
		{"invalid-mixed", "123abc", true, 42, 42},
		{"large-number", "9999999999", true, 42, 9999999999},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			envKey := "TEST_HP_VAR_" + tc.name

			if tc.setEnv {
				os.Setenv(envKey, tc.envValue)
				defer os.Unsetenv(envKey)
			} else {
				os.Unsetenv(envKey)
			}

			result := getEnvInt(envKey, tc.def)
			if result != tc.expected {
				t.Errorf("getEnvInt(%q=%q, %d) = %d; want %d",
					envKey, tc.envValue, tc.def, result, tc.expected)
			}
		})
	}
}
