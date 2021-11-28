package models

import "encoding/json"

type ServerConfig struct {
	ID   int    `yaml:"id" json:"id,omitempty"`
	Name string `yaml:"name" json:"name,omitempty"`
	Host string `yaml:"host" json:"host,omitempty"`
	Port int    `yaml:"port" json:"port,omitempty"`
}

func (config *ServerConfig) ToJson() ([]byte, error) {
	return json.Marshal(jsonDumpAble{
		Kind: config.GetKind(),
		Spec: config,
	})
}

func (config *ServerConfig) GetId() int {
	return config.ID
}

func (config *ServerConfig) SetId(id int) {
	config.ID = id
}

func (config *ServerConfig) GetName() string {
	return config.Name
}

func (config *ServerConfig) GetKind() string {
	return "ServerConfig"
}
