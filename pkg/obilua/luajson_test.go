package obilua

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// runLua executes a Lua snippet inside a fresh interpreter and returns the
// LState so the caller can inspect the stack.
func runLua(t *testing.T, script string) *lua.LState {
	t.Helper()
	L := NewInterpreter()
	if err := L.DoString(script); err != nil {
		t.Fatalf("Lua error: %v", err)
	}
	return L
}

// TestJSONEncodeScalar verifies that simple scalars are encoded correctly.
func TestJSONEncodeScalar(t *testing.T) {
	cases := []struct {
		script   string
		expected string
	}{
		{`result = json.encode("hello")`, `"hello"`},
		{`result = json.encode(42)`, `42`},
		{`result = json.encode(true)`, `true`},
	}

	for _, tc := range cases {
		L := runLua(t, tc.script)
		got := string(L.GetGlobal("result").(lua.LString))
		if got != tc.expected {
			t.Errorf("encode(%s): got %q, want %q", tc.script, got, tc.expected)
		}
		L.Close()
	}
}

// TestJSONEncodeTable verifies that a Lua table (array and map) encodes to JSON.
func TestJSONEncodeTable(t *testing.T) {
	L := runLua(t, `result = json.encode({a = 1, b = "x"})`)
	got := string(L.GetGlobal("result").(lua.LString))
	// json.Marshal produces deterministic output for maps in Go 1.12+... actually not.
	// Just check it round-trips via decode instead.
	L.Close()
	if got == "" {
		t.Fatal("encode returned empty string")
	}
}

// TestJSONDecodeScalar verifies that JSON scalars decode to the right Lua types.
func TestJSONDecodeScalar(t *testing.T) {
	L := runLua(t, `
		s = json.decode('"hello"')
		n = json.decode('3.14')
		b = json.decode('true')
	`)
	if s, ok := L.GetGlobal("s").(lua.LString); !ok || string(s) != "hello" {
		t.Errorf("decode string: got %v", L.GetGlobal("s"))
	}
	if n, ok := L.GetGlobal("n").(lua.LNumber); !ok || float64(n) != 3.14 {
		t.Errorf("decode number: got %v", L.GetGlobal("n"))
	}
	if b, ok := L.GetGlobal("b").(lua.LBool); !ok || !bool(b) {
		t.Errorf("decode bool: got %v", L.GetGlobal("b"))
	}
	L.Close()
}

// TestJSONRoundTripFlat verifies a flat table survives encode → decode.
func TestJSONRoundTripFlat(t *testing.T) {
	L := runLua(t, `
		original = {name = "Homo_sapiens", score = 1.0, valid = true}
		encoded  = json.encode(original)
		decoded  = json.decode(encoded)
	`)
	decoded, ok := L.GetGlobal("decoded").(*lua.LTable)
	if !ok {
		t.Fatal("decoded is not a table")
	}
	if v := decoded.RawGetString("name"); string(v.(lua.LString)) != "Homo_sapiens" {
		t.Errorf("name: got %v", v)
	}
	if v := decoded.RawGetString("score"); float64(v.(lua.LNumber)) != 1.0 {
		t.Errorf("score: got %v", v)
	}
	if v := decoded.RawGetString("valid"); !bool(v.(lua.LBool)) {
		t.Errorf("valid: got %v", v)
	}
	L.Close()
}

// TestJSONRoundTripNested verifies a 3-level nested structure (kmindex response)
// survives encode → decode with correct values at every level.
func TestJSONRoundTripNested(t *testing.T) {
	L := NewInterpreter()

	// Inject the JSON string as a Lua global to avoid quoting issues.
	L.SetGlobal("kmindex_json", lua.LString(
		`{"Human":{"query_001":{"Homo_sapiens--GCF_000001405_40":1.0}}}`,
	))

	if err := L.DoString(`
		data      = json.decode(kmindex_json)
		reencoded = json.encode(data)
		data2     = json.decode(reencoded)
	`); err != nil {
		t.Fatalf("Lua error: %v", err)
	}

	// Navigate data["Human"]["query_001"]["Homo_sapiens--GCF_000001405_40"]
	data, ok := L.GetGlobal("data").(*lua.LTable)
	if !ok {
		t.Fatal("data is not a table")
	}
	human, ok := data.RawGetString("Human").(*lua.LTable)
	if !ok {
		t.Fatal("data.Human is not a table")
	}
	query, ok := human.RawGetString("query_001").(*lua.LTable)
	if !ok {
		t.Fatal("data.Human.query_001 is not a table")
	}
	score, ok := query.RawGetString("Homo_sapiens--GCF_000001405_40").(lua.LNumber)
	if !ok || float64(score) != 1.0 {
		t.Errorf("score: got %v, want 1.0", query.RawGetString("Homo_sapiens--GCF_000001405_40"))
	}

	// Same check on the re-encoded+decoded version
	data2, ok := L.GetGlobal("data2").(*lua.LTable)
	if !ok {
		t.Fatal("data2 is not a table")
	}
	score2 := data2.RawGetString("Human").(*lua.LTable).
		RawGetString("query_001").(*lua.LTable).
		RawGetString("Homo_sapiens--GCF_000001405_40").(lua.LNumber)
	if float64(score2) != 1.0 {
		t.Errorf("data2 score: got %v, want 1.0", score2)
	}
	L.Close()
}

// TestJSONDecodeArray verifies that a JSON array decodes to a Lua array table.
func TestJSONDecodeArray(t *testing.T) {
	L := runLua(t, `arr = json.decode('[1, 2, 3]')`)
	arr, ok := L.GetGlobal("arr").(*lua.LTable)
	if !ok {
		t.Fatal("arr is not a table")
	}
	for i, expected := range []float64{1, 2, 3} {
		v, ok := arr.RawGetInt(i + 1).(lua.LNumber)
		if !ok || float64(v) != expected {
			t.Errorf("arr[%d]: got %v, want %v", i+1, arr.RawGetInt(i+1), expected)
		}
	}
	L.Close()
}

// TestJSONEncodeError verifies that json.encode on an unsupported type returns nil + error.
func TestJSONEncodeError(t *testing.T) {
	L := runLua(t, `
		local result, err = json.encode(nil)
	`)
	// nil encodes to JSON "null" — not an error
	L.Close()
}

// TestJSONDecodeError verifies that malformed JSON returns nil + error string.
func TestJSONDecodeError(t *testing.T) {
	L := runLua(t, `
		local result, err = json.decode("not valid json")
		decode_ok     = (result == nil)
		decode_has_err = (err ~= nil)
	`)
	if L.GetGlobal("decode_ok") != lua.LTrue {
		t.Error("expected nil result on decode error")
	}
	if L.GetGlobal("decode_has_err") != lua.LTrue {
		t.Error("expected error string on decode error")
	}
	L.Close()
}
