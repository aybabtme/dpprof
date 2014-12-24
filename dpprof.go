package dpprof

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aybabtme/dpprof/profile"
)

func CPU(w io.Writer, duration time.Duration, hosts ...string) error {
	form := url.Values{
		"seconds": {strconv.Itoa(int(duration.Seconds()))},
	}

	dumps, err := paraGet(nil, nil, form, hosts, "/debug/pprof/profile")
	if err != nil {
		return err
	}

	r := 1.0 / float64(len(dumps))

	var merger *profile.Profile
	for _, data := range dumps {
		prof, err := profile.Parse(bytes.NewBuffer(data))
		if err != nil {
			return err
		}
		if merger == nil {
			merger = prof
		} else if err := merger.Merge(prof, r); err != nil {
			return err
		}
	}

	return merger.Write(w)
}

func Symbol(r io.Reader, hosts ...string) (io.Reader, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	dumps, err := paraPost(nil, nil, nil, hosts, "/debug/pprof/symbol", data)
	if err != nil {
		return nil, err
	}

	// write the `num_symbols: 1` line first
	resp := bytes.NewBuffer(dumps[0])
	resp.WriteRune('\n')

	dedup := make(map[string]struct{})
	for _, dump := range dumps {
		str := string(dump)
		switch str {
		case string(dumps[0]), "":
			continue
		default:
		}
		if _, ok := dedup[str]; ok {
			continue
		}
		dedup[str] = struct{}{}
		_, _ = fmt.Fprintln(resp, str)
	}

	return resp, err
}

func NamedProfile(w io.Writer, name string, hosts ...string) error {

	dumps, err := paraGet(nil, nil, nil, hosts, "/debug/pprof/"+name)
	if err != nil {
		return err
	}

	r := 1.0 / float64(len(dumps))

	var merger *profile.Profile
	for _, data := range dumps {
		prof, err := profile.Parse(bytes.NewBuffer(data))
		if err != nil {
			if strings.Contains(err.Error(), "unrecognized profile format") {
				_, err = w.Write(data)
				return err
			}

			if strings.Contains(err.Error(), "malformed profile format") {
				_, err = w.Write(data)
				return err
			}

			log.Printf(string(bytes.Join(dumps, []byte{})))

			return err
		}
		if merger == nil {
			merger = prof
		} else if err := merger.Merge(prof, r); err != nil {
			return err
		}
	}
	return merger.Write(w)
}

func Cmdline(hosts ...string) ([][]byte, error) {
	return paraGet(nil, nil, nil, hosts, "/debug/pprof/cmdline")
}

func paraGet(c *http.Client, hdr http.Header, form url.Values, hosts []string, path string) ([][]byte, error) {
	if c == nil {
		c = http.DefaultClient
	}

	var wg sync.WaitGroup

	datac := make(chan []byte, len(hosts))
	errc := make(chan error, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			data, err := doGet(c, hdr, form, host, path)
			if err != nil {
				errc <- err
			} else {
				datac <- data
			}
			wg.Done()
		}(host)

	}
	wg.Wait()
	close(errc)
	close(datac)

	var out [][]byte
	for data := range datac {
		out = append(out, data)
	}
	return out, <-errc
}

func paraPost(c *http.Client, hdr http.Header, form url.Values, hosts []string, path string, body []byte) ([][]byte, error) {
	if c == nil {
		c = http.DefaultClient
	}

	var wg sync.WaitGroup

	datac := make(chan []byte, len(hosts))
	errc := make(chan error, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			data, err := doPost(c, hdr, form, host, path, bytes.NewReader(body))
			if err != nil {
				errc <- err
			} else {
				datac <- data
			}
			wg.Done()
		}(host)

	}
	wg.Wait()
	close(errc)
	close(datac)

	var out [][]byte
	for data := range datac {
		out = append(out, bytes.Split(data, []byte("\n"))...)
	}
	return out, <-errc
}

func doGet(c *http.Client, hdr http.Header, form url.Values, host string, path string) (data []byte, err error) {
	u, _ := url.Parse(host)
	u.Path = path
	q := u.Query()
	for k, vals := range form {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	for k, vals := range hdr {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	return do(c, req)
}

func doPost(c *http.Client, hdr http.Header, form url.Values, host string, path string, body io.Reader) (data []byte, err error) {
	u, _ := url.Parse(host)
	u.Path = path
	q := u.Query()
	for k, vals := range form {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, err
	}
	for k, vals := range hdr {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	return do(c, req)
}

func do(c *http.Client, req *http.Request) (data []byte, err error) {
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err = readError(resp); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func readError(r *http.Response) error {
	if r.StatusCode == 200 {
		return nil
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return &Error{Code: r.StatusCode, Msg: string(data)}
}

type Error struct {
	Code int
	Msg  string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Msg)
}
