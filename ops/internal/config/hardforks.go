package config

import (
	"fmt"
	"reflect"
	"time"
)

type HardforkTime uint64

func NewHardforkTime(t uint64) *HardforkTime {
	h := HardforkTime(t)
	return &h
}

func (h HardforkTime) MarshalTOML() ([]byte, error) {
	ts := time.Unix(int64(h), 0).UTC().Format("Mon 2 Jan 2006 15:04:05 MST")
	return []byte(fmt.Sprintf(`%d # %s`, uint64(h), ts)), nil
}

func (h *HardforkTime) U64Ptr() *uint64 {
	if h == nil {
		return nil
	}

	return (*uint64)(h)
}

func CopyHardforks(src *Hardforks, dst *Hardforks) error {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	// Iterate through all fields in the source struct
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)

		if srcField.IsNil() {
			continue
		}

		// Get the corresponding field in the destination
		dstField := dstVal.Field(i)

		// Only copy if dest is nil
		if !dstField.IsNil() {
			continue
		}

		// Copy the value
		dstField.Set(srcField)
	}

	return nil
}
