package science

import (
	"testing"

	"github.com/Clever/http-science/config"
	"github.com/stretchr/testify/assert"
)

func TestSimpleDiffs(t *testing.T) {
	eq := []byte("test string ';][23")
	neq := []byte("something else")

	assert.Equal(t, true, isSimplyEqual(eq, eq))
	assert.Equal(t, false, isSimplyEqual(eq, neq))
}

func TestJSONDiffs(t *testing.T) {

	for _, v := range []bool{true, false} {
		config.WeakCompare = v
		// Not equal when one or both are not json
		not := []byte("not json")
		jsonSimple2 := []byte(`{"yo": "jt"}`)
		jsonSimple1 := []byte(`{"yo": "jo"}`)
		assert.Equal(t, false, isJSONEqual(jsonSimple1, not))
		assert.Equal(t, false, isJSONEqual(not, not))

		// Correctly handles simple json
		assert.Equal(t, true, isJSONEqual(jsonSimple1, jsonSimple1))
		assert.Equal(t, false, isJSONEqual(jsonSimple1, jsonSimple2))

		// Correctly handles out of order keys
		jsonUnorderedKeys1 := []byte(`{"Hel": "lo", "Wor": "ld"}`)
		jsonUnorderedKeys2 := []byte(`{"Wor": "ld", "Hel": "lo"}`)
		assert.Equal(t, true, isJSONEqual(jsonUnorderedKeys1, jsonUnorderedKeys2))

		// Correctly handles nested objects
		jsonNested1 := []byte(`{"Hel": {"lo": ["Wor", "ld"]}}`)
		jsonNested2 := []byte(`{"Hel": {"lo": ["Wor", "ld!"]}}`)
		assert.Equal(t, true, isJSONEqual(jsonNested1, jsonNested1))
		assert.Equal(t, false, isJSONEqual(jsonNested1, jsonNested2))

		// We correctly handle duplicates in arrays
		jsonArrayDup1 := []byte(`{"Hello": ["Wor", "ld!", "ld!"]}}`)
		jsonArrayDup2 := []byte(`{"Hello": ["Wor", "Wor", "ld!"]}}`)
		assert.Equal(t, false, isJSONEqual(jsonArrayDup1, jsonArrayDup2))

		// Stress test
		assert.Equal(t, false, isJSONEqual(jsonComplicated, jsonComplicatedDifferent))
		assert.Equal(t, true, isJSONEqual(jsonComplicated, jsonComplicatedUnorderedKey))

		// OOO arrays and are different with weak vs strong comparison
		assert.Equal(t, v, isJSONEqual(jsonComplicated, jsonComplicatedUnorderedArray))
	}

}

var jsonComplicated = []byte(`
{
    "id": "0001",
    "type": "donut",
    "name": "Cake",
    "ppu": 0.55,
    "batters":
        {
            "batter":
                [
                    { "id": "1001", "type": "Regular" },
                    { "id": "1002", "type": "Chocolate" },
                    { "id": "1003", "type": "Blueberry" },
                    { "id": "1004", "type": "Devil's Food" }
                ]
        },
    "topping":
        [
            { "id": "5001", "type": "None" },
            { "id": "5002", "type": "Glazed" },
            { "id": "5005", "type": "Sugar" },
            { "id": "5007", "type": "Powdered Sugar" },
            { "id": "5006", "type": "Chocolate with Sprinkles" },
            { "id": "5003", "type": "Chocolate" },
            { "id": "5004", "type": ["Maple", "Chocolate"] }
        ]
}
`)

var jsonComplicatedUnorderedKey = []byte(`
{
    "id": "0001",
    "name": "Cake",
    "type": "donut",
    "ppu": 0.55,
    "topping":
        [
            { "id": "5001", "type": "None" },
            { "id": "5002", "type": "Glazed" },
            { "id": "5005", "type": "Sugar" },
            { "id": "5007", "type": "Powdered Sugar" },
            { "id": "5006", "type": "Chocolate with Sprinkles" },
            { "id": "5003", "type": "Chocolate" },
            { "id": "5004", "type": ["Maple", "Chocolate"] }
        ],
    "batters":
        {
            "batter":
                [
                    { "id": "1001", "type": "Regular" },
                    { "id": "1002", "type": "Chocolate" },
                    { "id": "1003", "type": "Blueberry" },
                    { "id": "1004", "type": "Devil's Food" }
                ]
        }
}
`)

var jsonComplicatedUnorderedArray = []byte(`
{
    "type": "donut",
    "id": "0001",
    "ppu": 0.55,
    "name": "Cake",
    "batters":
        {
            "batter":
                [
                    { "id": "1002", "type": "Chocolate" },
                    { "id": "1001", "type": "Regular" },
                    { "id": "1003", "type": "Blueberry" },
                    { "id": "1004", "type": "Devil's Food" }
                ]
        },
    "topping":
        [
            { "id": "5002", "type": "Glazed" },
            { "id": "5001", "type": "None" },
            { "id": "5007", "type": "Powdered Sugar" },
            { "id": "5005", "type": "Sugar" },
            { "id": "5003", "type": "Chocolate" },
            { "id": "5006", "type": "Chocolate with Sprinkles" },
            { "id": "5004", "type": ["Maple", "Chocolate"] }
        ]
}
`)

var jsonComplicatedDifferent = []byte(`
{
    "id": "0001",
    "type": "donut",
    "name": "Cake",
    "ppu": 0.55,
    "batters":
        {
            "batter":
                [
                    { "id": "1001", "type": "Regular" },
                    { "id": "1002", "type": "Chocolate" },
                    { "id": "1003", "type": "Blueberry" },
                    { "id": "1004", "type": "Devil's Food" }
                ]
        },
    "topping":
        [
            { "id": "5001", "type": "None" },
            { "id": "5002", "type": "Glazed" },
            { "id": "5005", "type": "Sugar" },
            { "id": "5007", "type": "Powdered Sugar" },
            { "id": "5006", "type": "Chocolate with Sprinkles" },
            { "id": "5003", "type": "DIFFERENT TOPPING" },
            { "id": "5004", "type": ["Maple", "Chocolate"] }
        ]
}
`)
