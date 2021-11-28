package models

import (
	"encoding/json"
	"ssh2/utils/tempfile"
)

const (
	AuthPassword            string = "PASSWORD"
	AUthPublishKeyFile      string = "PUBLISH_KEY_PATH"
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

func (auth *AuthMethod) ToJson() ([]byte, error) {
	return json.Marshal(jsonDumpAble{
		Kind: auth.GetKind(),
		Spec: auth,
	})
}

func (auth *AuthMethod) GetId() int {
	return auth.ID
}

func (auth *AuthMethod) SetId(id int) {
	auth.ID = id
}

func (auth *AuthMethod) GetName() string {
	return auth.Name
}

func (auth *AuthMethod) GetKind() string {
	return "AuthMethod"
}

func (auth *AuthMethod) GetPublishKeyPath() string {
	content := auth.GetDecryptedContent()
	if auth.Type == AuthPublishKey {
		file, err := tempfile.GetManager("").TempFile(auth.GetName())
		if err != nil {
			panic(err)
		}
		defer file.Close()
		file.WriteString(content)
		return file.Name()
		// TODO: 写到临时文件
		return content
	} else {
		return content
	}
}

func (auth *AuthMethod) GetDecryptedContent() string {
	return auth.Content
}
