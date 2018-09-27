package finder

import (
	"encoding/json"
	"fmt"
	"net/http"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/httputil"

	"git.xfyun.cn/AIaaS/finder-go/log"
)

func FeedbackForConfig(hc *http.Client, url string, f *common.ConfigFeedback) error {
	log.Log.Info("call FeedbackForConfig")
	contentType := "application/x-www-form-urlencoded"
	params := []byte(fmt.Sprintf("push_id=%s&project=%s&group=%s&service=%s&version=%s&addr=%s&config=%s&update_time=%d&update_status=%d&load_time=%d&load_status=%d&gray_group_id=%s",
		f.PushID, f.ServiceMete.Project, f.ServiceMete.Group, f.ServiceMete.Service, f.ServiceMete.Version, f.ServiceMete.Address,
		f.Config, f.UpdateTime, f.UpdateStatus, f.LoadTime, f.LoadStatus, f.GrayGroupId))
	result, err := httputil.DoPost(hc, contentType, url, params)
	if err != nil {
		log.Log.Error("FeedbackForConfig err:", err)
		err = errors.NewFinderError(errors.FeedbackPostErr)
		return err
	} else {
		log.Log.Info("FeedbackForConfig result:", string(result))
	}

	var r JSONResult
	err = json.Unmarshal([]byte(result), &r)
	if err != nil {
		log.Log.Error("FeedbackForConfig err:", err)
		err = errors.NewFinderError(errors.JsonUnmarshalErr)
		return err
	}
	if r.Ret != 0 {
		err = errors.NewFinderError(errors.FeedbackConfigErr)
		log.Log.Error("FeedbackForConfig err:", r.Msg)
		return err
	}

	return nil
}

func FeedbackForService(hc *http.Client, url string, f *common.ServiceFeedback) error {
	contentType := "application/x-www-form-urlencoded"
	params := []byte(fmt.Sprintf("push_id=%s&project=%s&group=%s&consumer=%s&consumer_version=%s&addr=%s&provider=%s&provider_version=%s&update_time=%d&update_status=%d&load_time=%d&load_status=%d&api_version=%s&type=%d",
		f.PushID, f.ServiceMete.Project, f.ServiceMete.Group, f.ServiceMete.Address, f.ServiceMete.Version, f.ServiceMete.Address,
		f.Provider, f.ProviderVersion, f.UpdateTime, f.UpdateStatus, f.LoadTime, f.LoadStatus, f.ProviderVersion, f.Type))
	result, err := httputil.DoPost(hc, contentType, url, params)
	if err != nil {
		log.Log.Error(err)
		err = errors.NewFinderError(errors.FeedbackPostErr)
		return err
	}else {
		log.Log.Info("FeedbackForService result:", string(result))
	}

	var r JSONResult
	err = json.Unmarshal([]byte(result), &r)
	if err != nil {
		log.Log.Error("[FeedbackForService][json]", err)
		err = errors.NewFinderError(errors.JsonUnmarshalErr)
		return err
	}
	if r.Ret != 0 {
		err = errors.NewFinderError(errors.FeedbackServiceErr)
		log.Log.Error("FeedbackServiceError :", r.Msg)
		return err
	}

	return nil
}
