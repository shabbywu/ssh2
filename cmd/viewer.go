package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/tidwall/buntdb"
	"github.com/urfave/cli/v2"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"ssh2/utils"
)

var db *buntdb.DB

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   string `json:"ttl,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

//go:embed viewer.html
var htmlTemplateData []byte
var htmlTemplate string

var viewerCommand = &cli.Command{Name: "viewer",
	Usage: "可视化UI",
	Action: func(context *cli.Context) (err error) {
		db, err = buntdb.Open(filepath.Join(utils.SSH2_HOME, "db.bin"))
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// 设置路由
		http.HandleFunc("/", indexHandler)
		http.HandleFunc("/api/keys", getKeysHandler)
		http.HandleFunc("/api/get", getKeyHandler)
		http.HandleFunc("/api/set", setKeyHandler)
		http.HandleFunc("/api/delete", deleteKeyHandler)

		port := ":8080"
		fmt.Printf("BuntDB Editor started at http://localhost%s\n", port)
		log.Fatal(http.ListenAndServe(port, nil))
		return nil
	},
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))
	tmpl.Execute(w, nil)
}

func getKeysHandler(w http.ResponseWriter, r *http.Request) {
	var keys []KeyValue

	err := db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(key, value string) bool {
			keys = append(keys, KeyValue{
				Key:   key,
				Value: value,
			})
			return true
		})
	})

	if err != nil {
		sendJSON(w, Response{Success: false, Message: err.Error()})
		return
	}

	sendJSON(w, Response{Success: true, Data: keys})
}

func getKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		sendJSON(w, Response{Success: false, Message: "Key is required"})
		return
	}

	var value string
	err := db.View(func(tx *buntdb.Tx) error {
		var err error
		value, err = tx.Get(key)
		return err
	})

	if err != nil {
		sendJSON(w, Response{Success: false, Message: err.Error()})
		return
	}

	sendJSON(w, Response{Success: true, Data: KeyValue{Key: key, Value: value}})
}

func setKeyHandler(w http.ResponseWriter, r *http.Request) {
	var kv KeyValue
	if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid JSON"})
		return
	}

	if kv.Key == "" {
		sendJSON(w, Response{Success: false, Message: "Key is required"})
		return
	}

	err := db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(kv.Key, kv.Value, nil)
		return err
	})

	if err != nil {
		sendJSON(w, Response{Success: false, Message: err.Error()})
		return
	}

	sendJSON(w, Response{Success: true, Message: "Key saved successfully"})
}

func deleteKeyHandler(w http.ResponseWriter, r *http.Request) {
	var kv KeyValue
	if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
		sendJSON(w, Response{Success: false, Message: "Invalid JSON"})
		return
	}

	if kv.Key == "" {
		sendJSON(w, Response{Success: false, Message: "Key is required"})
		return
	}

	err := db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(kv.Key)
		return err
	})

	if err != nil {
		sendJSON(w, Response{Success: false, Message: err.Error()})
		return
	}

	sendJSON(w, Response{Success: true, Message: "Key deleted successfully"})
}

func sendJSON(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func init() {
	htmlTemplate = string(htmlTemplateData)
	App.Commands = append(App.Commands, viewerCommand)
}
