package helpers

import (
    "testing"
    "time"
)


// Test generated using Keploy
func TestParseDateTimeTZ_ValidRFC3339WithTimezone(t *testing.T) {
    date := "2023-10-01T15:04:05+02:00"
    tz := time.UTC
    parsedTime, err := ParseDateTimeTZ(date, tz)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    expectedTime, _ := time.Parse(time.RFC3339, date)
    if !parsedTime.Equal(expectedTime) {
        t.Errorf("Expected %v, got %v", expectedTime, parsedTime)
    }
}


// Test generated using Keploy
func TestParseDateTimeTZ_ValidDateTimeWithoutTimezone(t *testing.T) {
    date := "2023-10-01 15:04:05"
    tz := time.FixedZone("UTC+2", 2*60*60)
    parsedTime, err := ParseDateTimeTZ(date, tz)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    expectedTime := time.Date(2023, 10, 1, 15, 4, 5, 0, tz)
    if !parsedTime.Equal(expectedTime) {
        t.Errorf("Expected %v, got %v", expectedTime, parsedTime)
    }
}


// Test generated using Keploy
func TestParseDateTimeTZ_ValidDateWithoutTime(t *testing.T) {
    date := "2023-10-01"
    tz := time.FixedZone("UTC+2", 2*60*60)
    parsedTime, err := ParseDateTimeTZ(date, tz)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    expectedTime := time.Date(2023, 10, 1, 0, 0, 0, 0, tz)
    if !parsedTime.Equal(expectedTime) {
        t.Errorf("Expected %v, got %v", expectedTime, parsedTime)
    }
}


// Test generated using Keploy
func TestFmtWakatimeDuration_ZeroDuration(t *testing.T) {
    duration := time.Duration(0)
    result := FmtWakatimeDuration(duration)
    expected := "0 hrs 0 mins"
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
