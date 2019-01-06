package listener

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestListenConfigs(t *testing.T) {
	wantOK := func(ctx context.Context, t *testing.T, ll ListenConfigs, network, address string) {
		t.Helper()
		l, err := ll.Listen(ctx, network, address)
		if err != nil {
			t.Errorf("%s, %s: unexpected error: %v", network, address, err)
			return
		}
		l.Close()
	}
	wantNG := func(ctx context.Context, t *testing.T, ll ListenConfigs, network, address string) {
		t.Helper()
		l, err := ll.Listen(ctx, network, address)
		if err != nil {
			return
		}
		l.Close()
		t.Errorf("%s, %s: error expected, got nil", network, address)
	}

	t.Run("ipv4", func(t *testing.T) {
		l, err := net.Listen("tcp4", ":0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()

		f, err := l.(interface{ File() (*os.File, error) }).File()
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, port, _ := net.SplitHostPort(l.Addr().String())

		ll := ListenConfigs{
			listenConfig{
				addr: l.Addr().String(),
				fd:   f.Fd(),
			},
		}
		wantOK(ctx, t, ll, "tcp", ":"+port)
		wantOK(ctx, t, ll, "tcp4", ":"+port)
		wantOK(ctx, t, ll, "tcp", "0.0.0.0:"+port)
		wantOK(ctx, t, ll, "tcp4", "0.0.0.0:"+port)
		wantNG(ctx, t, ll, "tcp6", ":"+port)
		wantNG(ctx, t, ll, "tcp", "[::]:"+port)
		wantNG(ctx, t, ll, "unix", "0.0.0.0:"+port)
	})

	t.Run("ipv4-loopback", func(t *testing.T) {
		l, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()

		f, err := l.(interface{ File() (*os.File, error) }).File()
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, port, _ := net.SplitHostPort(l.Addr().String())

		ll := ListenConfigs{
			listenConfig{
				addr: l.Addr().String(),
				fd:   f.Fd(),
			},
		}
		wantOK(ctx, t, ll, "tcp", "127.0.0.1:"+port)
		wantOK(ctx, t, ll, "tcp4", "127.0.0.1:"+port)
		wantOK(ctx, t, ll, "tcp", "localhost:"+port)
		wantOK(ctx, t, ll, "tcp4", "localhost:"+port)
		wantNG(ctx, t, ll, "tcp", "[::1]:"+port)
		wantNG(ctx, t, ll, "tcp6", "localhost:"+port)
		wantNG(ctx, t, ll, "unix", "127.0.0.1:"+port)
		wantNG(ctx, t, ll, "unix", "localhost:"+port)
	})

	t.Run("ipv6", func(t *testing.T) {
		l, err := net.Listen("tcp6", ":0")
		if err != nil {
			t.Skip("IPv6 is not supported?")
			return
		}
		defer l.Close()

		f, err := l.(interface{ File() (*os.File, error) }).File()
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, port, _ := net.SplitHostPort(l.Addr().String())

		ll := ListenConfigs{
			listenConfig{
				addr: l.Addr().String(),
				fd:   f.Fd(),
			},
		}
		wantOK(ctx, t, ll, "tcp", ":"+port)
		wantOK(ctx, t, ll, "tcp6", ":"+port)
		wantOK(ctx, t, ll, "tcp", "[::]:"+port)
		wantOK(ctx, t, ll, "tcp6", "[::]:"+port)
		wantNG(ctx, t, ll, "tcp", "0.0.0.0:"+port)
		wantNG(ctx, t, ll, "tcp4", ":"+port)
		wantNG(ctx, t, ll, "unix", ":"+port)
	})

	t.Run("ipv6-loopback", func(t *testing.T) {
		l, err := net.Listen("tcp6", "[::1]:0")
		if err != nil {
			t.Skip("IPv6 is not supported?")
			return
		}
		defer l.Close()

		f, err := l.(interface{ File() (*os.File, error) }).File()
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, port, _ := net.SplitHostPort(l.Addr().String())

		ll := ListenConfigs{
			listenConfig{
				addr: l.Addr().String(),
				fd:   f.Fd(),
			},
		}
		wantOK(ctx, t, ll, "tcp", "[::1]:"+port)
		wantOK(ctx, t, ll, "tcp6", "[::1]:"+port)
		wantOK(ctx, t, ll, "tcp", "localhost:"+port)
		wantOK(ctx, t, ll, "tcp6", "localhost:"+port)
		wantNG(ctx, t, ll, "tcp", "127.0.0.1:"+port)
		wantNG(ctx, t, ll, "tcp4", "localhost:"+port)
		wantNG(ctx, t, ll, "unix", "[::1]:"+port)
		wantNG(ctx, t, ll, "unix", "localhost:"+port)
	})

	t.Run("unix", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "server-starter-test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %s", err)
		}
		defer os.RemoveAll(dir)

		pwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("fail to getwd:%s", err)
		}
		os.Chdir(dir)
		defer os.Chdir(pwd)

		sock := filepath.Join(dir, "127.0.0.1:8000")
		l, err := net.Listen("unix", sock)
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()

		f, err := l.(interface{ File() (*os.File, error) }).File()
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ll := ListenConfigs{
			listenConfig{
				addr: "127.0.0.1:8000",
				fd:   f.Fd(),
			},
		}
		wantOK(ctx, t, ll, "unix", sock)
		wantOK(ctx, t, ll, "unix", "127.0.0.1:8000")
		wantNG(ctx, t, ll, "tcp", "127.0.0.1:8000")
		wantNG(ctx, t, ll, "tcp4", "127.0.0.1:8000")
	})
}

func TestPort(t *testing.T) {
	caces := []struct {
		in string
		ll []listenConfig
	}{
		{
			in: "0.0.0.0:80=3",
			ll: []listenConfig{
				{
					addr: "0.0.0.0:80",
					fd:   3,
				},
			},
		},
		{
			in: "0.0.0.0:80=3;/tmp/foo.sock=4",
			ll: []listenConfig{
				{
					addr: "0.0.0.0:80",
					fd:   3,
				},
				{
					addr: "/tmp/foo.sock",
					fd:   4,
				},
			},
		},
		{
			in: "50908=4",
			ll: []listenConfig{
				{
					addr: "50908",
					fd:   4,
				},
			},
		},
		{
			in: "",
			ll: []listenConfig{},
		},
	}

	for i, tc := range caces {
		ll, err := parseListenTargets(tc.in, true)
		if err != nil {
			t.Error(err)
			continue
		}
		if len(ll) != len(tc.ll) {
			t.Errorf("#%d: want %d, got %d", i, len(tc.ll), len(ll))
		}
		for i, l := range ll {
			l := l.(listenConfig)
			if !reflect.DeepEqual(l, tc.ll[i]) {
				t.Errorf("#%d, want %#v, got %#v", i, tc.ll[i], l)
			}
		}
		if ll.String() != tc.in {
			t.Errorf("#%d, want %s, got %s", i, tc.in, ll.String())
		}
	}

	errs := []string{
		"0.0.0.0:80=foo", // invalid fd
		"0.0.0.0:80",     // missing fd
	}
	for i, tc := range errs {
		ll, err := parseListenTargets(tc, true)
		if err == nil {
			t.Errorf("#%d: want error, got nil", i)
		}
		if ll != nil {
			t.Errorf("#%d: want nil, got %#v", i, ll)
		}
	}
}

func TestPortNoEnv(t *testing.T) {
	ports, err := parseListenTargets("", false)
	if err != ErrNoListeningTarget {
		t.Error("Ports must return error if no env")
	}

	if ports != nil {
		t.Errorf("Ports must return nil if no env")
	}
}
