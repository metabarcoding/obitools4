package obiformats

import (
	"bytes"
	"strconv"
	"strings"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/buger/jsonparser"
)

func _parse_json_map_string(str []byte) (map[string]string, error) {
	values := make(map[string]string)
	jsonparser.ObjectEach(str,
		func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) (err error) {
			skey := string(key)
			values[skey] = string(value)
			return
		},
	)
	return values, nil
}

func _parse_json_map_int(str []byte) (map[string]int, error) {
	values := make(map[string]int)
	jsonparser.ObjectEach(str,
		func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) (err error) {
			skey := string(key)
			intval, err := jsonparser.ParseInt(value)
			if err != nil {
				return err
			}
			values[skey] = int(intval)
			return nil
		},
	)
	return values, nil
}

func _parse_json_map_float(str []byte) (map[string]float64, error) {
	values := make(map[string]float64)
	jsonparser.ObjectEach(str,
		func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) (err error) {
			skey := string(key)
			floatval, err := strconv.ParseFloat(obiutils.UnsafeString(value), 64)
			if err != nil {
				return err
			}
			values[skey] = float64(floatval)
			return nil
		},
	)
	return values, nil
}

func _parse_json_map_bool(str []byte) (map[string]bool, error) {
	values := make(map[string]bool)
	jsonparser.ObjectEach(str,
		func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) (err error) {
			skey := string(key)
			boolval, err := jsonparser.ParseBoolean(value)
			if err != nil {
				return err
			}
			values[skey] = boolval
			return nil
		},
	)
	return values, nil
}

func _parse_json_map_interface(str []byte) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	jsonparser.ObjectEach(str,
		func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) (err error) {
			skey := string(key)
			switch dataType {
			case jsonparser.String:
				values[skey] = string(value)
			case jsonparser.Number:
				// Try to parse the number as an int at first then as float if that fails.
				values[skey], err = jsonparser.ParseInt(value)
				if err != nil {
					values[skey], err = strconv.ParseFloat(obiutils.UnsafeString(value), 64)
				}
				if err != nil {
					return
				}
			case jsonparser.Boolean:
			default:
				values[skey] = string(value)
			}
			return
		},
	)
	return values, nil
}

func _parse_json_array_string(str []byte) ([]string, error) {
	values := make([]string, 0)
	jsonparser.ArrayEach(str,
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if dataType == jsonparser.String {
				skey := string(value)
				values = append(values, skey)
			}
		},
	)
	return values, nil
}

func _parse_json_array_int(str []byte, sequence *obiseq.BioSequence) ([]int, error) {
	values := make([]int, 0)
	jsonparser.ArrayEach(str,
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if dataType == jsonparser.Number {
				intval, err := jsonparser.ParseInt(value)
				if err != nil {
					log.Fatalf("%s: Parsing int failed on value %s: %s", sequence.Id(), value, err)
				}
				values = append(values, int(intval))
			}
		},
	)
	return values, nil
}

func _parse_json_array_float(str []byte, sequence *obiseq.BioSequence) ([]float64, error) {
	values := make([]float64, 0)
	jsonparser.ArrayEach(str,
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if dataType == jsonparser.Number {
				floatval, err := strconv.ParseFloat(obiutils.UnsafeString(value), 64)
				if err == nil {
					values = append(values, float64(floatval))
				} else {
					log.Fatalf("%s: Parsing float failed on value %s: %s", sequence.Id(), value, err)
				}
			}
		},
	)
	return values, nil
}

func _parse_json_array_bool(str []byte, sequence *obiseq.BioSequence) ([]bool, error) {
	values := make([]bool, 0)
	jsonparser.ArrayEach(str,
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if dataType == jsonparser.Boolean {
				boolval, err := jsonparser.ParseBoolean(value)
				if err != nil {
					log.Fatalf("%s: Parsing bool failed on value %s: %s", sequence.Id(), value, err)
				}
				values = append(values, boolval)
			}
		},
	)
	return values, nil
}

func _parse_json_array_interface(str []byte) ([]interface{}, error) {
	values := make([]interface{}, 0)
	jsonparser.ArrayEach(str,
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			switch dataType {
			case jsonparser.String:
				values = append(values, string(value))
			case jsonparser.Number:
				// Try to parse the number as an int at first then as float if that fails.
				intval, err := jsonparser.ParseInt(value)
				if err != nil {
					floatval, err := strconv.ParseFloat(obiutils.UnsafeString(value), 64)
					if err != nil {
						values = append(values, string(value))
					} else {
						values = append(values, floatval)
					}
				} else {
					values = append(values, intval)
				}
			case jsonparser.Boolean:
				boolval, err := jsonparser.ParseBoolean(value)
				if err != nil {
					values = append(values, string(value))
				} else {
					values = append(values, boolval)
				}

			default:
				values = append(values, string(value))
			}

		},
	)
	return values, nil
}

func _parse_json_header_(header string, sequence *obiseq.BioSequence) string {
	annotations := sequence.Annotations()
	start := -1
	stop := -1
	level := 0
	lh := len(header)
	inquote := false

	for i := 0; (i < lh) && (stop < 0); i++ {
		// fmt.Printf("[%d,%d-%d] : %d (%c) (%d,%c)\n", i, start, stop, header[i], header[i], '{', '{')
		if level == 0 && header[i] == '{' && !inquote {
			start = i
		}

		// TODO: escaped double quotes are not considered
		if start > -1 && header[i] == '"' {
			inquote = !inquote
		}

		if header[i] == '{' && !inquote {
			level++
		}

		if header[i] == '}' && !inquote {
			level--
		}

		if start >= 0 && level == 0 {
			stop = i
		}

	}

	if start < 0 || stop < 0 {
		return header
	}

	stop++

	jsonparser.ObjectEach(obiutils.UnsafeBytes(header[start:stop]),
		func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			var err error

			skey := obiutils.UnsafeString(key)

			switch {
			case skey == "id":
				sequence.SetId(string(value))
			case skey == "definition":
				sequence.SetDefinition(string(value))

			case skey == "count":
				if dataType != jsonparser.Number {
					log.Fatalf("%s: Count attribut must be numeric: %s", sequence.Id(), string(value))
				}
				count, err := jsonparser.ParseInt(value)
				if err != nil {
					log.Fatalf("%s: Cannot parse count %s", sequence.Id(), string(value))
				}
				sequence.SetCount(int(count))

			case skey == "obiclean_weight":
				weight, err := _parse_json_map_int(value)
				if err != nil {
					log.Fatalf("%s: Cannot parse obiclean weight %s", sequence.Id(), string(value))
				}
				annotations[skey] = weight

			case skey == "obiclean_status":
				status, err := _parse_json_map_string(value)
				if err != nil {
					log.Fatalf("%s: Cannot parse obiclean status %s", sequence.Id(), string(value))
				}
				annotations[skey] = status

			case strings.HasPrefix(skey, "merged_"):
				if dataType == jsonparser.Object {
					data, err := _parse_json_map_int(value)
					if err != nil {
						log.Fatalf("%s: Cannot parse merged slot %s: %v", sequence.Id(), skey, err)
					} else {
						annotations[skey] = data
					}
				} else {
					log.Fatalf("%s: Cannot parse merged slot %s", sequence.Id(), skey)
				}

			case skey == "taxid":
				if dataType == jsonparser.Number || dataType == jsonparser.String {
					taxid := string(value)
					sequence.SetTaxid(taxid)
				} else {
					log.Fatalf("%s: Cannot parse taxid %s", sequence.Id(), string(value))
				}

			case strings.HasSuffix(skey, "_taxid"):
				if dataType == jsonparser.Number || dataType == jsonparser.String {
					rank, _ := obiutils.SplitInTwo(skey, '_')

					taxid := string(value)
					sequence.SetTaxid(taxid, rank)
				} else {
					log.Fatalf("%s: Cannot parse taxid %s", sequence.Id(), string(value))
				}

			default:
				skey = strings.Clone(skey)
				switch dataType {
				case jsonparser.String:
					annotations[skey] = string(value)
				case jsonparser.Number:
					// Try to parse the number as an int at first then as float if that fails.
					annotations[skey], err = jsonparser.ParseInt(value)
					if err != nil {
						annotations[skey], err = strconv.ParseFloat(obiutils.UnsafeString(value), 64)
					}
				case jsonparser.Array:
					annotations[skey], err = _parse_json_array_interface(value)
				case jsonparser.Object:
					annotations[skey], err = _parse_json_map_interface(value)
				case jsonparser.Boolean:
					annotations[skey], err = jsonparser.ParseBoolean(value)
				case jsonparser.Null:
					annotations[skey] = nil
				default:
					log.Fatalf("Unknown data type %v", dataType)
				}
			}

			if err != nil {
				annotations[skey] = "NaN"
				log.Fatalf("%s: Cannot parse value %s assicated to key %s into a %s value",
					sequence.Id(), string(value), skey, dataType.String())
			}

			return err
		},
	)

	// err := json.Unmarshal([]byte(header)[start:stop], &annotations)

	// for k, v := range annotations {
	// 	switch vt := v.(type) {
	// 	case float64:
	// 		if vt == math.Floor(vt) {
	// 			annotations[k] = int(vt)
	// 		}
	// 		{
	// 			annotations[k] = vt
	// 		}
	// 	}
	// }

	// if err != nil {
	// 	log.Fatalf("annotation parsing error on %s : %v\n", header, err)
	// }

	return strings.TrimSpace(header[stop:])
}

func ParseFastSeqJsonHeader(sequence *obiseq.BioSequence) {
	definition := sequence.Definition()
	sequence.SetDefinition("")

	definition_part := _parse_json_header_(
		definition,
		sequence,
	)

	if len(definition_part) > 0 {
		if sequence.HasDefinition() {
			definition_part = sequence.Definition() + " " + definition_part
		}
		sequence.SetDefinition(definition_part)
	}
}

func WriteFastSeqJsonHeader(buffer *bytes.Buffer, sequence *obiseq.BioSequence) {

	annotations := sequence.Annotations()

	if len(annotations) > 0 {
		err := obiutils.JsonMarshalByteBuffer(buffer, sequence.Annotations())

		if err != nil {
			log.Fatal(err)
		}
	}
}

func FormatFastSeqJsonHeader(sequence *obiseq.BioSequence) string {
	annotations := sequence.Annotations()
	buffer := bytes.Buffer{}

	if len(annotations) > 0 {
		obiutils.JsonMarshalByteBuffer(&buffer, sequence.Annotations())
		return unsafe.String(unsafe.SliceData(buffer.Bytes()), len(buffer.Bytes()))
	}

	return ""
}
