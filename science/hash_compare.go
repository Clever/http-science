package science

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/Clever/http-science/config"
)

////
//
// Sorting
//
////

// byCompare implements sort.Interface for md5 byte arrays
type byCompare [][md5.Size]byte

func (a byCompare) Len() int {
	return len(a)
}

func (a byCompare) Swap(i, j int) {
	for byteIndex, byteVal := range a[i] {
		tmp := a[j][byteIndex]
		a[j][byteIndex] = byteVal
		a[i][byteIndex] = tmp
	}
}

func (a byCompare) Less(i, j int) bool {
	return bytes.Compare(a[i][:], a[j][:]) == -1
}

////
//
// JSON Tree Hashing
//
////

func computeHashOfMd5Slice(md5s [][md5.Size]byte) [md5.Size]byte {
	data := make([]byte, 0, len(md5s)*md5.Size)
	for _, m := range md5s {
		data = append(data, m[:]...)
	}
	return md5.Sum(data)
}

func numberToHash(val interface{}) [md5.Size]byte {
	var buf bytes.Buffer
	// Pick any byte order
	binary.Write(&buf, binary.LittleEndian, val)
	return md5.Sum(buf.Bytes())
}

func computeHashForObject(jsonVal map[string]interface{}) ([md5.Size]byte, error) {
	n := len(jsonVal)
	keys := make([]string, 0, n)
	for k := range jsonVal {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	hashes := make([][md5.Size]byte, 0, n)
	for _, k := range keys {
		keyHash := md5.Sum([]byte(k))
		valHash, err := computeJSONTreeHash(jsonVal[k])
		if err != nil {
			return [md5.Size]byte{}, err
		}
		hashes = append(hashes, keyHash, valHash)
	}
	sort.Sort(byCompare(hashes))
	return computeHashOfMd5Slice(hashes), nil
}

func computeHashForArray(jsonVal []interface{}) ([md5.Size]byte, error) {
	n := len(jsonVal)
	hashes := make([][md5.Size]byte, 0, n)
	for _, v := range jsonVal {
		h, err := computeJSONTreeHash(v)
		if err != nil {
			return [md5.Size]byte{}, err
		}
		hashes = append(hashes, h)
	}
	sort.Sort(byCompare(hashes))
	return computeHashOfMd5Slice(hashes), nil
}

// computeJSONTreeHash computes a tree hash recursively for the JSON `jsonVal`,
// sorting contained arrays and objects so that they have the same hash
// regardless of the provided order.
func computeJSONTreeHash(jsonVal interface{}) ([md5.Size]byte, error) {
	var out [md5.Size]byte
	var err error

	switch jsonVal := jsonVal.(type) {
	case []interface{}:
		out, err = computeHashForArray(jsonVal)
	case map[string]interface{}:
		out, err = computeHashForObject(jsonVal)
	case string:
		out = md5.Sum([]byte(jsonVal))
	case int64:
		out = numberToHash(jsonVal)
	case float64:
		out = numberToHash(jsonVal)
	case bool:
		if jsonVal {
			out[0] = byte(1)
		} // else leave as default byte array
	default:
		// unknown field. possible explicit nil?
		msg := fmt.Sprintf("unable to decode json: %v\n", jsonVal)
		err = errors.New(msg)
	}
	return out, err
}

////
//
// JSON Object Comparison
//
////

// areEqualByHash returns true iff JSON objects `a` and `b` are equal by
// comparing their JSON tree hashes.
func areEqualByHash(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	hashA, errA := computeJSONTreeHash(a)
	hashB, errB := computeJSONTreeHash(b)
	if errA != nil || errB != nil {
		return false
	}

	return bytes.Compare(hashA[:], hashB[:]) == 0
}

func isJSONEqualHash(resControl, resExperiment []byte) bool {
	var controlJSON, expJSON map[string]interface{}
	// Return false if they can't be parsed as JSON
	if err := json.Unmarshal(resControl, &controlJSON); err != nil {
		return false
	}
	if err := json.Unmarshal(resExperiment, &expJSON); err != nil {
		return false
	}

	if config.WeakCompare {
		return areEqualByHash(controlJSON, expJSON)
	}
	return reflect.DeepEqual(controlJSON, expJSON)
}
