package main

import (
	"strings"
)

type ListFlag []string

func (l ListFlag) String() string {
	return strings.Join([]string(l), ",")
}

func (l *ListFlag) Set(v string) error {
	(*l) = append(*l, v)
	return nil
}

func (l ListFlag) Get() interface{} { return l }
