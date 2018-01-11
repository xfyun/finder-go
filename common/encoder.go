package finder

import (
	"encoding/binary"

	errors "git.xfyun.cn/AIaas/finder-go/errors"

	"git.xfyun.cn/AIaas/finder-go/utils/byteutil"
)

func EncodeValue(pushID string, data []byte) ([]byte, error) {
	var err error
	if len(data) == 0 {
		err = &errors.FinderError{
			Ret:  errors.InvalidParam,
			Func: "EncodeValue",
			Desc: "data is nil",
		}

		return nil, err
	}

	pushIDByte := []byte(pushID)
	pushIDLenByte := make([]byte, 4)
	binary.BigEndian.PutUint32(pushIDLenByte, uint32(len(pushIDByte)))

	return byteutil.Merge(pushIDLenByte, pushIDByte, data), nil
}
