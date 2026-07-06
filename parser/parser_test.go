package parser_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"os"
	"ssh2/db"
	"ssh2/models"
	"ssh2/parser"
)

var _ = Describe("Parser", func() {
	var data = `
---
kind: AuthMethod
spec:
  id: 2
  name: "方法"
  type: "PASSWORD"
  content: "内容"
  expect_for_password: "密码: "
---
kind: ClientConfig
spec:
  id: 1
  name: 测试1
  user: 用户A
  auth:
    spec:
      id: 3
      name: "内嵌方法"
      type: "PASSWORD"
      content: "内容"
      expect_for_password: "密码: "
`
	var decoder *yaml.Decoder
	BeforeEach(func() {
		decoder = yaml.NewDecoder(bytes.NewReader([]byte(data)))
	})

	It("test parse single record", func() {
		var record parser.DocumentRecord
		Expect(decoder.Decode(&record)).To(BeNil())
		ret, err := parser.YamlParser{}.ParseRecord(record)
		Expect(err).To(BeNil())

		Expect((*ret).GetKind()).To(Equal("AuthMethod"))
		Expect((*ret).GetId()).To(Equal(2))
		Expect((*ret).GetName()).To(Equal("方法"))

		jsonInDB, err := db.GetByField("AuthMethod", "id", 2)
		Expect(err).To(BeNil())

		jsonInRuntime, err := (*ret).ToJson()
		Expect(err).To(BeNil())
		Expect(jsonInDB).To(Equal(string(jsonInRuntime)))
	})

	It("test parse multi record", func() {
		var cnt int

		for _, expect := range []struct {
			Kind string
			Id   int
			Name string
		}{
			{Kind: "AuthMethod", Name: "方法", Id: 2},

			{Kind: "ClientConfig", Name: "测试1", Id: 1},
		} {

			var record parser.DocumentRecord
			Expect(decoder.Decode(&record)).To(BeNil())
			ret, err := parser.YamlParser{}.ParseRecord(record)
			Expect(err).To(BeNil())

			Expect((*ret).GetKind()).To(Equal(expect.Kind))
			Expect((*ret).GetId()).To(Equal(expect.Id))
			Expect((*ret).GetName()).To(Equal(expect.Name))

			cnt += 1
		}

		Expect(cnt).To(Equal(2))
	})

	It("encrypts parsed password auth methods", func() {
		ret, err := parser.YamlParser{}.ParseRecord(parser.DocumentRecord{
			Kind: "AuthMethod",
			Spec: map[interface{}]interface{}{
				"id":                  20,
				"name":                "encrypted-password",
				"type":                "PASSWORD",
				"content":             "secret",
				"expect_for_password": "password:",
			},
		})
		Expect(err).To(BeNil())

		auth := (*ret).(*models.AuthMethod)
		Expect(auth.Content).NotTo(Equal("secret"))
		content, err := auth.DecryptedContent()
		Expect(err).To(BeNil())
		Expect(content).To(Equal("secret"))
	})

	It("supports interactive password auth methods", func() {
		ret, err := parser.YamlParser{}.ParseRecord(parser.DocumentRecord{
			Kind: "AuthMethod",
			Spec: map[interface{}]interface{}{
				"id":                  21,
				"name":                "interactive-password",
				"type":                "INTERACTIVE_PASSWORD",
				"expect_for_password": "password:",
			},
		})
		Expect(err).To(BeNil())

		auth := (*ret).(*models.AuthMethod)
		Expect(auth.Type).To(Equal(models.AUthInteractivePassword))
		content, err := auth.DecryptedContent()
		Expect(err).To(BeNil())
		Expect(content).To(Equal(""))
	})

	It("stores private key file content when requested", func() {
		keyFile, err := os.CreateTemp("", "ssh2-test-key")
		Expect(err).To(BeNil())
		defer os.Remove(keyFile.Name())
		_, err = keyFile.WriteString("PRIVATE KEY")
		Expect(err).To(BeNil())
		Expect(keyFile.Close()).To(BeNil())

		ret, err := parser.YamlParser{}.ParseRecord(parser.DocumentRecord{
			Kind: "AuthMethod",
			Spec: map[interface{}]interface{}{
				"id":                     22,
				"name":                   "stored-key",
				"type":                   "PUBLISH_KEY_PATH",
				"content":                keyFile.Name(),
				"save_private_key_in_db": true,
			},
		})
		Expect(err).To(BeNil())

		auth := (*ret).(*models.AuthMethod)
		Expect(auth.Type).To(Equal(models.AuthPublishKey))
		content, err := auth.DecryptedContent()
		Expect(err).To(BeNil())
		Expect(content).To(Equal("PRIVATE KEY"))
	})
})
