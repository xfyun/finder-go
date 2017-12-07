package zkutil

import (
	"encoding/binary"
	"finder-go/common"
	"finder-go/errors"
)

func DecodeValue(data []byte) (string, []byte, error) {
	var err error
	if len(data) == 0 {
		err = &errors.FinderError{
			Ret:  common.InvalidParam,
			Func: "decodeValue",
			Desc: "data is nil",
		}

		return "", nil, err
	}
	if len(data) <= 4 {
		err = &errors.FinderError{
			Ret:  common.InvalidParam,
			Func: "decodeValue",
			Desc: "len of data < =4",
		}

		return "", nil, err
	}
	l := binary.BigEndian.Uint32(data[:4])
	if int(l) > (len(data) - 4) {
		err = &errors.FinderError{
			Ret:  common.InvalidParam,
			Func: "decodeValue",
			Desc: "len of data <= 4",
		}

		return "", nil, err
	}
	pushId := string(data[4 : l+4])

	return pushId, data[l+4:], nil
}
