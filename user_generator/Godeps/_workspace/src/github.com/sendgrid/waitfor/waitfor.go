package waitfor

/*
Use the waitfor package to retry something until it succeeds.
*/

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type TimeoutError struct {
	t time.Duration //timeout
	m string        // message
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("waitfor: timed out waiting %s %s", e.t.String(), e.m)
}

func newTimeoutError(timeout time.Duration, additionalMessage string) TimeoutError {
	return TimeoutError{t: timeout, m: additionalMessage}
}

// Func takes a function which takes no args (nullary) and returns an error,
// and will block and repeatedly call that function on the specified interval until it
// does not return an error. If the timeout is reached, an error will be
// returned.
func Func(f func() error, interval, timeout time.Duration) error {
	if interval > timeout {
		panic(fmt.Sprintf(
			"waitfor: interval (%s) must be shorter than timeout (%s)",
			interval.String(), timeout.String()))
	}

	lastError := error(nil)
	errChan := make(chan error)
	timeoutChan := time.After(timeout)

	wrapper := func() {
		select {
		case errChan <- f():
		default:
		}
	}

	for {
		// Run f() in a goroutine so we can still timeout if it takes a long
		// time
		go wrapper()

		select {

		case <-timeoutChan:
			// Get the last error if it's there. If it's not, it means the function never
			// returned within the timeout period
			if lastError != nil {
				return newTimeoutError(timeout, fmt.Sprintf("; last error: '%s'", lastError))
			}

			return newTimeoutError(timeout, "; function did not return")

		case err := <-errChan:
			if err == nil {
				return nil
			}

			if err != nil {
				lastError = err
				time.Sleep(interval)
			}
		}
	}
}

// HTTP will attempt to GET the specified url every interval until the timeout
// is reached. If an HTTP 200 response is not returned within the timeout
// period, an error will be returned.
func HTTP(url string, interval, timeout time.Duration) error {
	return Func(func() error {
		rsp, err := http.Get(url)
		if err != nil {
			return err
		}

		if rsp.StatusCode != 200 {
			return fmt.Errorf("non-200 status for GET %s: %s", url, rsp.Status)
		}

		return nil
	}, interval, timeout)
}

// Dial will attempt to connect to the specified network/address until the
// timeout is reached. An error will be returned if the address cannot be
// reached within the timeout period.
func Dial(network, address string, interval, timeout time.Duration) error {
	return Func(func() error {
		conn, err := net.DialTimeout(network, address, interval)
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}, 0, timeout)
}
