package models

import (
	"encoding/json"
)

type ClientConfig struct {
	ID           int    `yaml:"id" json:"id,omitempty"`
	Name         string `yaml:"name" json:"name,omitempty"`
	User         string `yaml:"user" json:"user,omitempty"`
	AuthMethodId int    `yaml:"auth_method_id" json:"auth_method_id,omitempty"`
}

func (config *ClientConfig) ToJson() ([]byte, error) {
	return json.Marshal(jsonDumpAble{
		Kind: config.GetKind(),
		Spec: config,
	})
}

func (config *ClientConfig) GetAuthMethod() (*AuthMethod, error) {
	ptr, err := GetByField[AuthMethod]("AuthMethod", "id", config.AuthMethodId)
	if err != nil {
		return nil, err
	}
	return &ptr, nil
}

func (config *ClientConfig) GetId() int {
	return config.ID
}

func (config *ClientConfig) SetId(id int) {
	config.ID = id
}

func (config *ClientConfig) GetName() string {
	return config.Name
}

func (config *ClientConfig) GetKind() string {
	return "ClientConfig"
}
