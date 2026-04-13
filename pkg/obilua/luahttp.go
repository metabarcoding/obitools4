package obilua

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	lua "github.com/yuin/gopher-lua"
)

const httpClientTimeout = 30 * time.Second

var (
	_httpClient     *http.Client
	_httpClientOnce sync.Once
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
	})
	return _httpClient
}

// RegisterHTTP registers the http module in the Lua state as a global,
// consistent with obicontext and BioSequence.
//
// Exposes:
//
//	http.post(url, body) → response string  (on success)
//	http.post(url, body) → nil, err string  (on error)
func RegisterHTTP(luaState *lua.LState) {
	table := luaState.NewTable()
	luaState.SetField(table, "post", luaState.NewFunction(luaHTTPPost))
	luaState.SetGlobal("http", table)
}

// luaHTTPPost implements http.post(url, body) for Lua.
//
// Lua signature:
//
//	local response = http.post(url, body)
//	local response, err = http.post(url, body)
func luaHTTPPost(L *lua.LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := getHTTPClient().Do(req)
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
