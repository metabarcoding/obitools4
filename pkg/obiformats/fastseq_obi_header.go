package obiformats

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/goccy/go-json"
)

var __obi_header_value_string_pattern__ = regexp.MustCompile(`^'\s*([^']*'|"[^"]*")\s*;`)
var __obi_header_value_numeric_pattern__ = regexp.MustCompile(`^\s*([+-]?\.\d+|[+-]?\d+(\.\d*)?([eE][+-]?\d+)?)\s*;`)
var __obi_header_map_int_key__ = regexp.MustCompile("([{,])([0-9]+):")

func __match__dict__(text []byte) []int {

	state := 0
	level := 0
	start := 0
	instring := byte(0)

	for i, r := range text {
		if state == 2 {
			if r == ';' {
				// end of the pattern
				return []int{start, i + 1}
			}

			if r != ' ' && r != '\t' {
				// Bad character at the end of the pattern
				return []int{}
			}
		}

		if r == '{' && instring == 0 { // Beginning of dict
			level++
			if state == 0 {
				// Beginning of the main dict
				state++
				start = i
			}

			continue
		}

		if state == 0 && r != ' ' && r != '\t' {
			// It's not a dict
			return []int{}
		}

		if state == 1 {
			if r == '"' || r == '\'' {
				if instring == 0 {
					// start of a string
					instring = r
				} else {
					if instring == r {
						// end of a string
						instring = 0
					}
				}

				continue
			}
		}

		if r == '}' && instring == 0 {
			// end of a dict
			level--

			if level == 0 {
				// end of the main dict
				state++
			}
		}

	}

	return []int{}
}

func __match__key__(text []byte) []int {

	state := 0
	start := 0

	for i, r := range text {

		if state == 0 {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				// Beginning of the key
				// fmt.Printf("Beginning of the key (%c) %d\n", r, i)
				state++
				start = i
				continue
			}

			if r != ' ' && r != '\t' {
				// It's not a key
				return []int{}
			}

			continue
		}

		if state > 0 && r == '=' {
			// End of thee pattern
			// fmt.Printf("End of the pattern (%c) %d\n", r, i)
			return []int{start, i + 1}
		}

		if state == 1 {
			if r == ' ' || r == '\t' {
				// End of the key
				state++
				continue
			}

			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
				(r >= '0' && r <= '9') ||
				r == '_' || r == '-' || r == '.' {
				// Continuing the key
				continue
			}

			// Not allowed character in a key
			// fmt.Printf("Not allowed char (%c) %d\n", r, i)
			return []int{}
		}

		if state == 2 && r != ' ' && r != '\t' {
			// fmt.Printf("Not allowed char 2 (%c) %d\n", r, i)
			// Not allowed character after a key
			return []int{}
		}
	}

	return []int{} // Not a key
}

func __match__general__(text []byte) []int {

	for i, r := range text {
		if r == ';' {
			return []int{0, i + 1}
		}
	}

	return []int{} // Not generic value
}

var __false__ = []byte{'f', 'a', 'l', 's', 'e'}
var __False__ = []byte{'F', 'a', 'l', 's', 'e'}
var __FALSE__ = []byte{'F', 'A', 'L', 'S', 'E'}

var __true__ = []byte{'t', 'r', 'u', 'e'}
var __True__ = []byte{'T', 'r', 'u', 'e'}
var __TRUE__ = []byte{'T', 'R', 'U', 'E'}

func __is_true__(text []byte) bool {
	return (len(text) == 1 && (text[0] == 't' || text[0] == 'T')) ||
		bytes.Equal(text, __true__) ||
		bytes.Equal(text, __True__) ||
		bytes.Equal(text, __TRUE__)
}

func __is_false__(text []byte) bool {
	return (len(text) == 1 && (text[0] == 'f' || text[0] == 'F')) ||
		bytes.Equal(text, __false__) ||
		bytes.Equal(text, __False__) ||
		bytes.Equal(text, __FALSE__)
}

func ParseOBIFeatures(text string, annotations obiseq.Annotation) string {

	definition := []byte(text)
	d := definition

	for m := __match__key__(definition); len(m) > 0; {
		var bvalue []byte
		var value interface{}
		start := m[0]
		stop := -1
		key := string(bytes.TrimSpace(d[start:(m[1] - 1)]))
		part := d[m[1]:]

		// numeric value
		m = __obi_header_value_numeric_pattern__.FindIndex(part)
		if len(m) > 0 {
			bvalue = bytes.TrimSpace(part[m[0]:(m[1] - 1)])
			value, _ = strconv.ParseFloat(string(bvalue), 64)
			stop = m[1] + 1
		} else {
			// string value

			m = __obi_header_value_string_pattern__.FindIndex(part)
			if len(m) > 0 {
				bvalue = bytes.TrimSpace(part[m[0]:(m[1] - 1)])
				value = string(bvalue[1:(len(bvalue) - 1)])
				stop = m[1] + 1
			} else {

				// dict value
				m = __match__dict__(part)
				if len(m) > 0 {
					bvalue = bytes.TrimSpace(part[m[0]:(m[1] - 1)])
					j := bytes.ReplaceAll(bvalue, []byte("'"), []byte(`"`))
					j = __obi_header_map_int_key__.ReplaceAll(j, []byte(`$1"$2":`))
					var err error
					switch {
					case strings.HasSuffix(key, "_count"):
						dict := make(map[string]int)
						err = json.Unmarshal(j, &dict)
						value = dict
					case strings.HasPrefix(key, "merged_"):
						dict := make(map[string]int)
						err = json.Unmarshal(j, &dict)
						value = obiseq.MapAsStatsOnValues(dict)
					case strings.HasSuffix(key, "_status") ||
						strings.HasSuffix(key, "_mutation"):
						dict := make(map[string]string)
						err = json.Unmarshal(j, &dict)
						value = dict

					default:
						dict := make(map[string]interface{})
						err = json.Unmarshal(j, &dict)
						value = dict
					}

					if err != nil {
						value = string(bvalue)
					}
					stop = m[1] + 1
				} else {

					// Generic value

					// m = __obi_header_value_general_pattern__.FindIndex(part)
					m = __match__general__(part)
					if len(m) > 0 {
						bvalue = bytes.TrimSpace(part[m[0]:(m[1] - 1)])

						if __is_false__(bvalue) {
							value = false
						} else {
							if __is_true__(bvalue) {
								value = true
							} else {
								value = string(bvalue)
							}
						}

						stop = m[1] + 1
					} else {
						// no value
						break
					} // End of No value
				} // End of not dict
			} // End of not string
		} // End of not numeric

		switch vt := value.(type) {
		case float64:
			if vt == math.Floor(vt) {
				annotations[key] = int(vt)
			}
		default:
			annotations[key] = value
		}

		if stop < len(part) {
			d = part[stop:]
		} else {
			d = []byte{}
		}
		//m = __obi_header_key_pattern__.FindIndex(d)
		m = __match__key__(d)
	}

	return string(bytes.TrimSpace(d))
}

func ParseFastSeqOBIHeader(sequence *obiseq.BioSequence) {
	annotations := sequence.Annotations()
	definition := sequence.Definition()
	sequence.SetDefinition("")

	definition = ParseOBIFeatures(definition, annotations)

	if len(definition) > 0 {
		if sequence.HasDefinition() {
			definition = sequence.Definition() + " " + definition
		}
		sequence.SetDefinition(definition)
	}
}

func WriteFastSeqOBIHeade(buffer *bytes.Buffer, sequence *obiseq.BioSequence) {

	annotations := sequence.Annotations()

	if len(annotations) > 0 {

		for key, value := range annotations {
			if key != "definition" {

				switch t := value.(type) {
				case string:
					buffer.WriteString(fmt.Sprintf("%s=%s; ", key, t))
				case *obiseq.StatsOnValues:
					t.RLock()
					tv, err := obiutils.JsonMarshal(t.Map())
					t.RUnlock()
					if err != nil {
						log.Fatalf("Cannot convert %v value", value)
					}
					tv = bytes.ReplaceAll(tv, []byte(`"`), []byte("'"))
					buffer.WriteString(fmt.Sprintf("%s=", key))
					buffer.Write(tv)
					buffer.WriteString("; ")
				case map[string]int,
					map[string]string,
					map[string]interface{}:
					tv, err := obiutils.JsonMarshal(t)
					if err != nil {
						log.Fatalf("Cannot convert %v value", value)
					}
					tv = bytes.ReplaceAll(tv, []byte(`"`), []byte("'"))
					buffer.WriteString(fmt.Sprintf("%s=", key))
					buffer.Write(tv)
					buffer.WriteString("; ")
				default:
					buffer.WriteString(fmt.Sprintf("%s=%v; ", key, value))
				}
			}
		}

		if sequence.HasDefinition() {
			buffer.WriteByte(' ')
			buffer.WriteString(sequence.Definition())
		}
	}

}

func FormatFastSeqOBIHeader(sequence *obiseq.BioSequence) string {
	annotations := sequence.Annotations()

	if annotations != nil {
		var text bytes.Buffer
		WriteFastSeqOBIHeade(&text, sequence)
		return unsafe.String(unsafe.SliceData(text.Bytes()), len(text.String()))
	}

	return ""
}
