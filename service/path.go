package service

import (
	"youfile/config"
	"youfile/youplus"
)

func GetRealPath(target string, token string) (string, error) {
	if config.Instance.YouPlusPath {
		realPath, err := youplus.DefaultClient.GetRealPath(target, token)
		if err != nil {
			return "", err
		}
		return realPath, nil
	}
	return target, nil
}
