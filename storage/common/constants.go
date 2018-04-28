package common

type DataChangedCallback func(path string, data []byte) error
