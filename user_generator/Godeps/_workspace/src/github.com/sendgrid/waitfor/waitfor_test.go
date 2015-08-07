package waitfor

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type funcTestCase struct {
	f                 func() error
	label             string
	interval, timeout time.Duration
	expectedErr       string
}

func TestFunc(t *testing.T) {
	t.Parallel()

	start := time.Now()

	cases := []funcTestCase{
		{
			label: "Finishes the first time",
			f: func() error {
				return nil
			},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: "",
		},

		{
			label: "Fails a few times but finishes before the timeout",
			f: func() error {
				if time.Now().After(start.Add(300 * time.Millisecond)) {
					return fmt.Errorf("test error")
				}

				return nil
			},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: "",
		},

		{
			label: "Times out",
			f: func() error {
				return fmt.Errorf("won't finish")
			},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: "won't finish",
		},

		{
			label: "Times out even if f never returns",
			f: func() error {
				time.Sleep(1 * time.Hour)
				return fmt.Errorf("won't return")
			},
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: "function did not return",
		},
	}

	var wg sync.WaitGroup
	for i, c := range cases {
		wg.Add(1)
		go func(j int, c funcTestCase) {
			err := Func(c.f, c.interval, c.timeout)
			if c.expectedErr != "" && !strings.Contains(err.Error(), c.expectedErr) {
				t.Errorf("#%d %s: got error %q, expected %q", j, c.label, err, c.expectedErr)
			}
			wg.Done()
		}(i, c)

	}

	wg.Wait()
}

func TestFuncPanic(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil || !strings.Contains(fmt.Sprint(e), "must be shorter than timeout") {
			t.Errorf("got panic '%v', expected: %q", e, "must be shorter than timeout")
		}
	}()

	Func(func() error { return nil }, time.Hour, time.Second)
	t.Errorf("expected panic, got none")
}

type httpTestCase struct {
	url               string
	interval, timeout time.Duration
	expectedErr       string
}

func TestHTTP(t *testing.T) {
	t.Parallel()

	// Set up a test HTTP server which responds to GET /?code=xxx with a status code of xxx
	s := httptest.NewServer(http.DefaultServeMux)
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		code, err := strconv.Atoi(req.FormValue("code"))
		if err != nil {
			code = 500
		}

		rw.WriteHeader(code)
	})
	defer s.Close()

	cases := []httpTestCase{
		{
			url:         s.URL + "/?code=200",
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: ""},
		{
			url:         s.URL + "/?code=400",
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: "400"},
	}

	var wg sync.WaitGroup
	for i, c := range cases {
		wg.Add(1)
		go func(j int, c httpTestCase) {
			err := HTTP(c.url, c.interval, c.timeout)
			if c.expectedErr != "" && !strings.Contains(err.Error(), c.expectedErr) {
				t.Errorf("#%d %q: got error %q, expected %q", j, c.url, err, c.expectedErr)
			}
			if c.expectedErr == "" && err != nil {
				t.Errorf("#%d %q: expected no error, got %q", j, c.url, err)
			}
			wg.Done()
		}(i, c)
	}

	wg.Wait()
}

type tcpTestCase struct {
	addr              string
	interval, timeout time.Duration
	expectedErr       string
}

func TestTCP(t *testing.T) {
	t.Parallel()

	// Set up a test TCP server
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		for {
			conn, err := l.Accept()

			if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			conn.Close()
		}
	}()

	cases := []tcpTestCase{
		{
			addr:        l.Addr().String(),
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: ""},
		{
			addr:        "1.2.3.4:5",
			interval:    100 * time.Millisecond,
			timeout:     1 * time.Second,
			expectedErr: "timeout"},
	}

	var wg sync.WaitGroup
	for i, c := range cases {
		wg.Add(1)
		go func(j int, c tcpTestCase) {
			err := Dial("tcp", c.addr, c.interval, c.timeout)
			if c.expectedErr != "" && !strings.Contains(err.Error(), c.expectedErr) {
				t.Errorf("#%d %q: got error %q, expected %q", j, c.addr, err, c.expectedErr)
			}
			if c.expectedErr == "" && err != nil {
				t.Errorf("#%d %q: expected no error, got %q", j, c.addr, err)
			}
			wg.Done()
		}(i, c)
	}

	wg.Wait()
}
