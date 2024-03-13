package main

import (
	"bytes"
	"errors"
	"strings"
)

type conv uint8

const (
	UC conv = 1 << iota
	LC
	TR
)

func (f *conv) String() string {
	var ss []string
	if f.has(UC) {
		ss = append(ss, "upper_case")
	}
	if f.has(LC) {
		ss = append(ss, "lower_case")
	}
	if f.has(TR) {
		ss = append(ss, "trim_spaces")
	}
	return strings.Join(ss, ",")
}

func (f *conv) Set(s string) error {
	ss := strings.Split(s, ",")
	for _, v := range ss {
		switch v {
		case "upper_case":
			f.set(UC)
		case "lower_case":
			f.set(LC)
		case "trim_spaces":
			f.set(TR)
		default:
			return errors.New("unknown flag")
		}
	}
	if f.has(UC) && f.has(LC) {
		return errors.New("upper_case and lower_case both set")
	}
	return nil
}

func (f *conv) set(flag conv) {
	*f |= flag
}

func (f *conv) has(flag conv) bool {
	return (*f)&flag != 0
}

func convert(b []byte, conv conv) []byte {
	if conv.has(TR) {
		b = bytes.TrimSpace(b)
	}
	if conv.has(LC) {
		b = bytes.ToLower(b)
	} else if conv.has(UC) {
		b = bytes.ToUpper(b)
	}

	return b
}
