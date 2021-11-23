package parser_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"ssh2/db"
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
})
