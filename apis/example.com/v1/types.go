package v1

import (
	"encoding/json"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/runtime/schema"
	meta_v1 "k8s.io/client-go/pkg/apis/meta/v1"
)

type HelloWorld struct {
	meta_v1.TypeMeta `json:",inline"`
	Metadata             api.ObjectMeta `json:"metadata"`

	Spec HelloWorldSpec `json:"spec"`
}

type HelloWorldSpec struct {
	Foo string `json:"foo"`
	Bar bool   `json:"bar"`
}

type HelloWorldList struct {
	meta_v1.TypeMeta `json:",inline"`
	Metadata             meta_v1.ListMeta `json:"metadata"`

	Items []HelloWorld `json:"items"`
}

// Required to satisfy Object interface
func (e *HelloWorld) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (e *HelloWorld) GetObjectMeta() meta.Object {
	return &e.Metadata
}

// Required to satisfy Object interface
func (el *HelloWorldList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (el *HelloWorldList) GetListMeta() meta.List {
	return &el.Metadata
}

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

type HelloWorldListCopy HelloWorldList
type HelloWorldCopy HelloWorld

func (e *HelloWorld) UnmarshalJSON(data []byte) error {
	tmp := HelloWorldCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := HelloWorld(tmp)
	*e = tmp2
	return nil
}

func (el *HelloWorldList) UnmarshalJSON(data []byte) error {
	tmp := HelloWorldListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := HelloWorldList(tmp)
	*el = tmp2
	return nil
}
