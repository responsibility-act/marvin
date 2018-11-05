package chores

import (
	"bytes"
	"encoding/json"
	"github.com/bigblind/v-eight"
	"github.com/gobuffalo/packr"
	"github.com/marvin-automator/marvin/actions"
	"github.com/marvin-automator/marvin/internal/db"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"text/template"
)


type RegisteredTrigger struct {
	Provider string
	Group    string
	Action   string
}

type ConfigInput struct {
	Name        string
	Description string
}

type ChoreTemplateConfig struct {
	Triggers []RegisteredTrigger
	Inputs   []ConfigInput
}

type ChoreTemplate struct {
	Name   string
	Id     string
	Script string
	Config ChoreTemplateConfig
}

var (
	errTemplateNotFound = errors.New("Template not found")
)

func NewChoreTemplate(name, script string) (*ChoreTemplate, error) {
	ct := ChoreTemplate{
		Name: name,
		Script: script,
		Id: uuid.NewV4().String(),
	}

	err := ct.GenerateTemplateConfigs()
	return &ct, err
}

var templateCache map[string]*ChoreTemplate
var cacheLoaded = false
const storeName = "chore_templates"

func LoadChoreTemplate(id string) (*ChoreTemplate, error) {
	ct, ok := templateCache[id]
	if !ok {
		return ct, errTemplateNotFound
	}

	s := db.GetStore(storeName)
	ct = new(ChoreTemplate)
	err := s.Get(id, ct)
	if _, ok := err.(db.KeyNotFoundError); ok {
		return ct, errTemplateNotFound
	}

	if err != nil {
		return ct, err
	}

	templateCache[id] = ct
	return ct, nil
}

func GetChoreTemplates() ([]*ChoreTemplate, error) {
	results := make([]*ChoreTemplate, 0, len(templateCache))

	if cacheLoaded {
		for _, ct := range templateCache {
			results = append(results, ct)
		}
		return results, nil
	}

	s := db.GetStore(storeName)
	ct := new(ChoreTemplate)
	err := s.EachKeyWithPrefix("", ct, func(key string) error {
		newPtr := &(*ct)
		templateCache[key] = newPtr
		results = append(results, newPtr)
		return nil
	})

	if err != nil {
		return nil, err
	}

	cacheLoaded = true
	return results, nil
}

func (ct *ChoreTemplate) Save() error {
	s := db.GetStore("chore_templates")
	templateCache[ct.Id] = ct
	return s.Set(ct.Id, ct)
}

var bp, bpErr = packr.NewBox("./js").FindString("boilerplate.js.template")
var bpTemplate = template.Must(template.New("js").Parse(bp))

func (ct *ChoreTemplate) combineScriptWithBoilerplate(inputs map[string]string) string {
	if bpErr != nil {
		panic(bpErr) // If everything is set up correctly during build, this shouldn't happen.
	}

	w := bytes.NewBuffer([]byte{})
	bpTemplate.Execute(w, struct{
		Providers []actions.Provider
		Inputs map[string]string
	}{actions.Registry.Providers(), inputs})

	s := w.String() + ct.Script
	return s
}

func (ct *ChoreTemplate) GetChoreSnapshot(inputs map[string]string) []byte {
	s := v8.CreateSnapshot(ct.combineScriptWithBoilerplate(inputs))
	return s.Export()
}

func (ct *ChoreTemplate) getConfigV8Value(inputs map[string]string) (*v8.Value, error) {
	is := v8.NewIsolate()
	return is.NewContext().Eval(ct.combineScriptWithBoilerplate(inputs)+";marvin;", "marvin.js")

}

func (ct *ChoreTemplate) GenerateTemplateConfigs() error {
	res, err := ct.getConfigV8Value(nil)
	if err != nil {
		return err
	}

	ct.Config = ChoreTemplateConfig{}

	err = decodeFieldValue(res, "_triggers", &(ct.Config.Triggers))
	if err != nil {
		return err
	}

	return decodeFieldValue(res, "_inputs", &(ct.Config.Inputs))
}

func (ct *ChoreTemplate) GenerateChoreConfig(inputValues map[string]string) (*choreConfig, error) {
	res, err := ct.getConfigV8Value(inputValues)
	if err != nil {
		return nil, err
	}

	cc := choreConfig{}
	cc.Inputs = inputValues

	err = decodeFieldValue(res, "_triggers", &(cc.Triggers))
	return &cc, err
}

func decodeFieldValue(value *v8.Value, field string, ptr interface{}) error {
	v, err := value.Get(field)
	if err != nil {
		return err
	}

	data, err := v.MarshalJSON()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, ptr)
}
