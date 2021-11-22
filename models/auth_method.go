package models

import "encoding/json"

const (
	AuthPassword string = "PASSWORD"
	AUthPublishKeyFile string = "PUBLISH_KEY_FILE"
	AuthPublishKey string = "PUBLISH_KEY"
	AUthInteractivePassword string = "INTERACTIVE_PASSWORD"
)

type AuthMethod struct {
	ID int
	Name string
	Type string
	Content string
	ExpectForPassword string
}


func (auth *AuthMethod) ToJson () ([]byte, error){
	return json.Marshal(jsonDumpAble{
		Kind: "AuthMethod",
		Spec: auth,
	})
}

func (auth AuthMethod) GetId() int {
	return auth.ID
}

func (auth AuthMethod) GetName() string {
	return auth.Name
}

func (auth AuthMethod) GetKind() string {
	return "AuthMethod"
}
