package filterlist

import (
	"errors"
	"testing"
	"time"
)

func TestRetrierSuccess(t *testing.T) {
	expectedContent := "foo"
	sleeper := MockSleeper{}
	retrier := Retrier{
		fetcher: &MockFetcher{content: expectedContent, remainingFailurs: 0},
		sleeper: &sleeper,
	}
	res, err := retrier.FetchWithRetryAndBackoff(5, time.Minute*5, 10)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if res != expectedContent {
		t.Errorf("Expected content to be '%s', but got '%s'", expectedContent, res)
	}
	if sleeper.SecondsWaited() != 0 {
		t.Errorf("Expected retrier to sleep 0 seconds, but got %v", sleeper.SecondsWaited())
	}
}

func TestRetryWithoutBackoff(t *testing.T) {
	expectedContent := "foo"
	expectedWait := 0
	sleeper := MockSleeper{}
	retrier := Retrier{
		fetcher: &MockFetcher{content: expectedContent, remainingFailurs: 4},
		sleeper: &sleeper,
	}
	res, err := retrier.FetchWithRetryAndBackoff(5, time.Minute*5, 10)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if res != expectedContent {
		t.Errorf("Expected content to be '%s', but got '%s'", expectedContent, res)
	}
	if sleeper.SecondsWaited() != expectedWait {
		t.Errorf("Expected retrier to sleep %v seconds, but got %v", expectedWait, sleeper.SecondsWaited())
	}
}

func TestRetryWithBackoff(t *testing.T) {
	expectedContent := "foo"
	expectedWait := 60 * 5 // 5 Minutes
	sleeper := MockSleeper{}
	retrier := Retrier{
		fetcher: &MockFetcher{content: expectedContent, remainingFailurs: 6},
		sleeper: &sleeper,
	}
	res, err := retrier.FetchWithRetryAndBackoff(5, time.Minute*5, 10)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if res != expectedContent {
		t.Errorf("Expected content to be '%s', but got '%s'", expectedContent, res)
	}
	if sleeper.SecondsWaited() != expectedWait {
		t.Errorf("Expected retrier to sleep %v seconds, but got %v", expectedWait, sleeper.SecondsWaited())
	}
}

func TestRetryMultipleBackoffs(t *testing.T) {
	expectedContent := "foo"
	expectedWait := 60 * 10 // 10 Minutes
	sleeper := MockSleeper{}
	retrier := Retrier{
		fetcher: &MockFetcher{content: expectedContent, remainingFailurs: 11},
		sleeper: &sleeper,
	}
	res, err := retrier.FetchWithRetryAndBackoff(5, time.Minute*5, 20)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if res != expectedContent {
		t.Errorf("Expected content to be '%s', but got '%s'", expectedContent, res)
	}
	if sleeper.SecondsWaited() != expectedWait {
		t.Errorf("Expected retrier to sleep %v seconds, but got %v", expectedWait, sleeper.SecondsWaited())
	}
}

func TestRetryReachingFailureThreshold(t *testing.T) {
	expectedWait := 60 * 5 // 5 Minutes
	sleeper := MockSleeper{}
	retrier := Retrier{
		fetcher: &MockFetcher{content: "foo", remainingFailurs: 5},
		sleeper: &sleeper,
	}
	_, err := retrier.FetchWithRetryAndBackoff(5, time.Minute*5, 5)
	if !errors.Is(err, ErrFailureCountReached) {
		t.Errorf("Expected error wrapping FailureCountReachedError, but got %v", err)
	}
	if sleeper.SecondsWaited() != expectedWait {
		t.Errorf("Expected retrier to sleep %v seconds, but got %v", expectedWait, sleeper.SecondsWaited())
	}
}
