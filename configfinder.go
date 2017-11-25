package finder

import (
	"finder-go/errors"
)

type Config struct {
	name    string
	file    []byte
	oldFile []byte
}

type ConfigFinder struct {

}

type OnCfgUpdateEvent func(Config) int

func (f *ConfigFinder) UseConfig(name string, dynamic bool, event OnCfgUpdateEvent) error {
	err := new(errors.FinderError)
	return err
}

func (f *ConfigFinder) DestroyConfig(name string) error {
	return nil
}
