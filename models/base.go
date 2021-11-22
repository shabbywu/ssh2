package models

import (
	"encoding/json"
	"fmt"
	"github.com/thedevsaddam/gojsonq/v2"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type Model interface {
	GetId() int
	GetName() string
	GetKind() string
}

//jsonDumpAble: 将模型序列化到 json(数据库)
type jsonDumpAble struct {
	Kind string
	Spec interface{}
}

type Ref struct {
	Field string
	Value interface{}
}

var key string

var db *gojsonq.JSONQ
var cache interface{}

func init()  {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	key = path.Join(home, ".ssh/data.json")

	db = gojsonq.New().File(key)
	cache = db.Get()
}

func Save() {
	data, err := json.Marshal(cache)
	if err != nil {
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(key, data, 0666); err != nil{
		log.Fatal(err)
		return
	}
}

func Get(kind string, ref *Ref) (interface{}){
	if ref == nil {
		return db.Where("kind", "=", kind).Get()
	}
	return db.Where("kind", "=", kind).Where(fmt.Sprintf("spec.%s", ref.Field), "=", ref.Value).Get()
}
