package finder

import (
	"encoding/json"

	"fmt"
	"net/http"

	errors "git.xfyun.cn/AIaas/finder-go/errors"
	"git.xfyun.cn/AIaas/finder-go/utils/httputil"
)

func RegisterService(hc *http.Client, url string, project string, group string, service string) error {
	contentType := "application/x-www-form-urlencoded"
	params := []byte(fmt.Sprintf("project=%s&group=%s&service=%s", project, group, service))
	result, err := httputil.DoPost(hc, contentType, url, params)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var r JSONResult
	err = json.Unmarshal([]byte(result), &r)
	if err != nil {
		return err
	}
	if r.Ret != 0 {
		err = &errors.FinderError{
			Ret:  errors.FeedbackServiceError,
			Func: "RegisterService",
			Desc: r.Msg,
		}

		return err
	}

	return nil
}
