/*
* CODE GENERATED AUTOMATICALLY WITH github.com/ernesto-jimenez/gogen/unmarshalmap
* THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package testpkg

import (
	"fmt"
)

// UnmarshalMap takes a map and unmarshals the fieds into the struct
func (s *Array) UnmarshalMap(m map[string]interface{}) error {

	// Array List

	if v, ok := m["List"].([]string); ok {
		s.List = make([]string, len(v))
		for i, el := range v {
			s.List[i] = el
		}
	} else if v, exists := m["List"]; exists && v != nil {
		return fmt.Errorf("expected field List to be []string but got %T", m["List"])
	}

	return nil
}
