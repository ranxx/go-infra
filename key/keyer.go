package key

import (
	"strings"
)

var defaultSep = ":"

// Keyer key
type Keyer interface {
	Key(e ...string) string
}

// defaultKeyer keyer
var defaultKeyer = &keyer{sep: defaultSep}

type keyer struct {
	sep string
	p   string
}

// NewKeyer new keyer
func NewKeyer(sep string, e ...string) *keyer {
	return &keyer{
		sep: sep,
		p:   (&keyer{sep: sep}).Key(e...),
	}
}

func InitKeyer(sep string, prefix ...string) *keyer {
	defaultKeyer = NewKeyer(sep, prefix...)
	return defaultKeyer
}

// Key key
func (k *keyer) Key(e ...string) string {
	return k.internalWithPrefix(k.p, k.sep, e...)
}

func (k *keyer) internalWithPrefix(prefix, sep string, e ...string) string {
	slice := make([]string, 0, len(e)+1)
	if len(prefix) > 0 {
		slice = append(slice, prefix)
	}
	if len(e) > 0 {
		slice = append(slice, e...)
	}
	if len(slice) <= 0 {
		return ""
	}
	return strings.Join(slice, sep)
}

// WrapKey wrap keyer
type WrapKey func(e ...string) string

// Key key
func (w WrapKey) Key(e ...string) string {
	return w(e...)
}

// Key ...
func Key(e ...string) string {
	return defaultKeyer.Key(e...)
}

// SimpleKey ...
func SimpleKey(name string) Keyer {
	return WrapKey(func(e ...string) string {
		return name
	})
}
