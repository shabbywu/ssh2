package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"ssh2/utils/crypto"
	"ssh2/utils/tempfile"
)

const (
	AuthPassword            string = "PASSWORD"
	AUthPublishKeyFile      string = "PUBLISH_KEY_PATH"
	AuthPublishKey          string = "PUBLISH_KEY_CONTENT"
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

func (auth *AuthMethod) PublishKeyPath() (string, func(), error) {
	content, err := auth.DecryptedContent()
	if err != nil {
		return "", nil, err
	}
	if auth.Type == AuthPublishKey {
		file, err := tempfile.GetManager("").TempFile(auth.GetName())
		if err != nil {
			return "", nil, err
		}
		defer file.Close()
		if o, err := base64.StdEncoding.DecodeString(content); err == nil {
			if _, err := file.Write(o); err != nil {
				return "", nil, err
			}
		} else {
			if _, err := file.WriteString(content); err != nil {
				return "", nil, err
			}
		}
		if err := os.Chmod(file.Name(), 0600); err != nil {
			return "", nil, err
		}
		return file.Name(), func() { _ = os.Remove(file.Name()) }, nil
	}
	if auth.Type != AUthPublishKeyFile {
		return "", nil, fmt.Errorf("auth type %s does not provide a publish key", auth.Type)
	}
	path := filepath.Clean(content)
	if err := os.Chmod(path, 0600); err != nil {
		return "", nil, err
	}
	return path, func() {}, nil
}

func (auth *AuthMethod) GetPublishKeyPath() string {
	path, _, err := auth.PublishKeyPath()
	if err != nil {
		panic(err)
	}
	return path
}

func (auth *AuthMethod) EncryptContent() error {
	encrypted, err := crypto.Encrypt(auth.Content)
	if err != nil {
		return err
	}
	auth.Content = encrypted
	return nil
}

func (auth *AuthMethod) DecryptedContent() (string, error) {
	return crypto.Decrypt(auth.Content)
}

func (auth *AuthMethod) GetDecryptedContent() string {
	content, err := auth.DecryptedContent()
	if err != nil {
		return auth.Content
	}
	return content
}
