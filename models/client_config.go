package models

import (
	"encoding/json"
	"errors"
)

type ClientConfig struct {
	ID int
	Name string
	User string
	AuthMethodId int
}

func (config *ClientConfig) ToJson () ([]byte, error){
	return json.Marshal(jsonDumpAble{
		Kind: "ClientConfig",
		Spec: config,
	})
}

func (config *ClientConfig) GetAuthMethod() (*AuthMethod, error) {
	spec := db.Where("kind", "=", "AuthMethod").Where("spec.id", "=", config.AuthMethodId).Select("spec").Get()
	obj, ok := spec.(AuthMethod)
	if !ok {
		return nil, errors.New("非法的 AuthMethod 结构体")
	}
	return &obj, nil
}

func (config ClientConfig) GetId() int {
	return config.ID
}

func (config ClientConfig) GetName() string {
	return config.Name
}

func (config ClientConfig) GetKind() string {
	return "ClientConfig"
}
