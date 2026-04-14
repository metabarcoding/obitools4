package obilua

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	lua "github.com/yuin/gopher-lua"
)

const httpClientTimeout = 300 * time.Second

var (
	_httpClient     *http.Client
	_httpClientOnce sync.Once

	// _httpSemaphore limits the number of concurrent HTTP requests.
	// Initialised lazily alongside the client.
	_httpSemaphore chan struct{}
)

func getHTTPClient() *http.Client {
	_httpClientOnce.Do(func() {
		conns := 2 * obidefault.ParallelWorkers()
		_httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: conns,
				MaxConnsPerHost:     conns,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: httpClientTimeout,
		}
		_httpSemaphore = make(chan struct{}, obidefault.ParallelWorkers())
	})
	return _httpClient
}

// RegisterHTTP registers the http module in the Lua state as a global,
// consistent with obicontext and BioSequence.
//
// Exposes:
//
//	http.post(url, body [, timeout_ms]) → response string  (on success)
//	http.post(url, body [, timeout_ms]) → nil, err string  (on error)
//	http.set_concurrency(n)             → set max simultaneous requests
func RegisterHTTP(luaState *lua.LState) {
	table := luaState.NewTable()
	luaState.SetField(table, "post", luaState.NewFunction(luaHTTPPost))
	luaState.SetField(table, "set_concurrency", luaState.NewFunction(luaHTTPSetConcurrency))
	luaState.SetGlobal("http", table)
}

// luaHTTPPost implements http.post(url, body [, timeout_ms]) for Lua.
//
// The optional third argument overrides the default timeout (in milliseconds).
// Concurrent requests are throttled through _httpSemaphore so that a
// single-threaded backend server is not overwhelmed by K parallel workers.
//
// Lua signature:
//
//	local response          = http.post(url, body)
//	local response          = http.post(url, body, 5000)   -- 5 s timeout
//	local response, err     = http.post(url, body)
func luaHTTPPost(L *lua.LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)

	client := getHTTPClient()

	timeout := httpClientTimeout
	if L.GetTop() >= 3 {
		ms := L.CheckInt(3)
		timeout = time.Duration(ms) * time.Millisecond
	}

	// Acquire semaphore slot — blocks until a slot is free.
	_httpSemaphore <- struct{}{}
	defer func() { <-_httpSemaphore }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(respBytes))
	return 1
}

// luaHTTPSetConcurrency replaces the semaphore with a new one of size n.
// Must be called before the first http.post (e.g. in begin()).
//
// Lua signature:
//
//	http.set_concurrency(1)   -- serialise all HTTP requests
func luaHTTPSetConcurrency(L *lua.LState) int {
	n := L.CheckInt(1)
	if n < 1 {
		n = 1
	}
	getHTTPClient() // ensure singleton is initialised
	_httpSemaphore = make(chan struct{}, n)
	return 0
}
