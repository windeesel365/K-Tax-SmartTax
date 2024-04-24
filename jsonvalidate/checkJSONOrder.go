package jsonvalidate

import (
	"fmt"
	"strings"
)

// checkJSONOrder checks if the keys in the provided JSON body are in the expected order.
func CheckJSONOrder(body []byte, expectedKeys []string) error {

	keys, err := GetOrderedKeysFromJSON(body)
	if err != nil {
		fmt.Println("Error:", err)
	}

	keys = keys[0:len(expectedKeys)]

	expectedKeysMembers := strings.Join(expectedKeys, ", ")

	if len(keys) != len(expectedKeys) {
		return fmt.Errorf("please enter just %d key(s) ordered by: %s. Then process again", len(expectedKeys), expectedKeysMembers)
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			//if key != expectedKeys[i] && !unexpectedform {
			return fmt.Errorf("please enter data in correct order, key name: %s. Then process again", expectedKeysMembers)
		}
	}

	// clear keys after processing
	keys = nil

	return nil
}
