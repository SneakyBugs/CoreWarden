package filterlist

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var FailureCountReachedError = fmt.Errorf("Fetch error failure count reached")

type Retrier struct {
	fetcher Fetcher
	sleeper Sleeper
}

func (r Retrier) FetchWithRetryAndBackoff(failuresUntilBackoff int, backoffWait time.Duration, failuresUntilError int) (result string, err error) {
	remainingUntilBackoff := failuresUntilBackoff
	for failures := 0; failures < failuresUntilError; failures++ {
		res, err := r.fetcher.Fetch()
		if err == nil {
			return res, nil
		}
		remainingUntilBackoff--
		if remainingUntilBackoff == 0 {
			remainingUntilBackoff = failuresUntilBackoff
			r.sleeper.Sleep(backoffWait)
		}
	}
	return "", fmt.Errorf("%w: failed %d times", FailureCountReachedError, failuresUntilError)
}

type Fetcher interface {
	Fetch() (string, error)
}

type URLFetcher struct {
	url string
}

func (f URLFetcher) Fetch() (string, error) {
	resp, err := http.Get(f.url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	return string(body), nil
}

type MockFetcher struct {
	content          string
	remainingFailurs int
}

func (f *MockFetcher) Fetch() (string, error) {
	if 0 < f.remainingFailurs {
		f.remainingFailurs--
		return "", fmt.Errorf("mock error")
	}
	return f.content, nil
}

type Sleeper interface {
	Sleep(duration time.Duration)
}

type RealSleeper struct{}

func (s RealSleeper) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

type MockSleeper struct {
	waited time.Duration
}

func (s *MockSleeper) Sleep(duration time.Duration) {
	s.waited += duration
}

func (s MockSleeper) SecondsWaited() int {
	return int(s.waited.Seconds())
}
