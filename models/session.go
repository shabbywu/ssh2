package models

import (
	"encoding/json"
)

type Session struct {
	ID             int    `yaml:"id" json:"id,omitempty"`
	Name           string `yaml:"name" json:"name,omitempty"`
	Tag            string `yaml:"tag" json:"tag,omitempty"`
	ClientConfigId int    `yaml:"client_config_id" json:"client_config_id,omitempty"`
	ServerConfigId int    `yaml:"server_config_id" json:"server_config_id,omitempty"`

	Plugins string `yaml:"plugins" json:"plugins,omitempty"`
}

func (s *Session) ToJson() ([]byte, error) {
	return json.Marshal(jsonDumpAble{
		Kind: s.GetKind(),
		Spec: s,
	})
}

func (s *Session) GetClientConfig() (*ClientConfig, error) {
	ptr, err := GetByField[ClientConfig]("ClientConfig", "id", s.ClientConfigId)
	if err != nil {
		return nil, err
	}
	return &ptr, nil
}

func (s *Session) GetServerConfig() (*ServerConfig, error) {
	ptr, err := GetByField[ServerConfig]("ServerConfig", "id", s.ServerConfigId)
	if err != nil {
		return nil, err
	}
	return &ptr, nil
}

func (s *Session) GetId() int {
	return s.ID
}

func (s *Session) SetId(id int) {
	s.ID = id
}

func (s *Session) GetName() string {
	return s.Name
}

func (s *Session) GetKind() string {
	return "Session"
}
