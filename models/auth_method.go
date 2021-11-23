package models

import "encoding/json"

const (
	AuthPassword            string = "PASSWORD"
	AUthPublishKeyFile      string = "PUBLISH_KEY_FILE"
	AuthPublishKey          string = "PUBLISH_KEY"
	AUthInteractivePassword string = "INTERACTIVE_PASSWORD"
)

type AuthMethod struct {
	ID                int    `yaml:"id" json:"id,omitempty"`
	Name              string `yaml:"name" json:"name,omitempty"`
	Type              string `yaml:"type" json:"type,omitempty"`
	Content           string `yaml:"content" json:"content,omitempty"`
	ExpectForPassword string `yaml:"expect_for_password" json:"expect_for_password,omitempty"`
}

func (auth AuthMethod) ToJson() ([]byte, error) {
	return json.Marshal(jsonDumpAble{
		Kind: auth.GetKind(),
		Spec: auth,
	})
}

func (auth AuthMethod) GetId() int {
	return auth.ID
}

func (auth AuthMethod) SetId(id int) {
	auth.ID = id
}

func (auth AuthMethod) GetName() string {
	return auth.Name
}

func (auth AuthMethod) GetKind() string {
	return "AuthMethod"
}
