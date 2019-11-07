package mulbase

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

var basicFields = MakeFieldHolder([]string{
	"value",
	"value2",
	"value6",
})

const (
	_valueType      = "Value"
	_innerValueType = "InnerValue"
)

var rootUid string

var defaultFields = basicFields.Copy().Append(basicFields, "value3|").Append(basicFields, "value4")

func TestMain(m *testing.M) {
	setup()
	//mutateSetup()
	code := m.Run()
	os.Exit(code)
}

type Value struct {
	BaseNode
	StringVal      string     `json:"value"`
	IntVal         int        `json:"value2"`
	UidVal         InnerValue `json:"value3"`
	UidArrayVal    []Value    `json:"value4"`
	IntArrayVal    []int      `json:"value5"`
	StringArrayVal []string   `json:"value6"`
}

type InnerValue struct {
	BaseNode
	InnerInt   int    `json:"value7"`
	FacetValue string `json:"value3|facetValue,omitempty"`
}

func (i *InnerValue) GetAllInfo() map[string]interface{} {
	m := i.GetRelativeInfo()
	m["value7"] = i.InnerInt
	return m
}

func (i *InnerValue) GetRelativeInfo() map[string]interface{} {
	m := i.BaseNode.GetRelativeInfo()
	if i.FacetValue != "" {
		m["value3|facetValue"] = i.FacetValue
	}
	return m
}

func (i *InnerValue) SetType() {
	i.AddType(_innerValueType)
}

func (v *Value) SetType() {
	if v.Uid != "" {
		//Do not set type for an existing node.
		return
	}
	v.AddType(_valueType)
	v.UidVal.SetType()
	for i := 0; i < len(v.UidArrayVal); i++ {
		v.UidArrayVal[i].SetType()
	}
}

func (v *Value) GetAllInfo() map[string]interface{} {
	m := make(map[string]interface{})
	err := StructToMap(v, &m)
	if err != nil {
		return nil
	}
	return m
}

func (v *Value) GetRelativeInfo() map[string]interface{} {
	m := makeUIDMap(v.Uid)
	return m
}

func (v *Value) UID() string {
	return v.Uid
}

func (v *Value) DeleteUIDS() map[string]interface{} {
	m := makeUIDMap(v.Uid)
	m["value3"] = v.UidVal.DeleteUIDS()
	var arr = make([]map[string]interface{}, len(v.UidArrayVal))
	for i := 0; i < len(v.UidArrayVal); i++ {
		arr[i] = v.UidArrayVal[i].DeleteUIDS()
	}
	m["value4"] = arr
	b, _ := json.Marshal(m)
	fmt.Println(string(b))
	return m
}

func mutateSetup() {
	v := Value{}
	v.IntVal = 5
	v.StringVal = "Test"
	v.IntArrayVal = []int{1, 3, 5}
	v.StringArrayVal = []string{"S", "W", "A", "G"}
	s := v
	for i := 0; i < 30; i++ {
		temp := s
		temp.StringVal = temp.StringVal + strconv.Itoa(i)
		temp.IntVal = i
		temp.UidVal.InnerInt = 5
		v.UidArrayVal = append(v.UidArrayVal, temp)
	}
	v.UidVal.InnerInt = 7
	m, err := CreateDNode(&v, true)
	rootUid = GetRootUID(m)
	if err != nil {
		panic(err)
	}
}

func setup() {
	_test = true
	GraphInit("127.0.0.1", 9080, false)
}

func TestEmptyQuery(t *testing.T) {
	q := NewQuery()
	err := q.Execute(nil)
	if err == nil {
		t.Fail()
	}
}

func TestFind(t *testing.T) {
	var v Value
	err := Find(&v, rootUid, defaultFields)
	if err != nil || v.StringVal == "" {
		t.Log(err)
		t.Fail()
	}

}

func TestType(t *testing.T) {
	var v Value
	var Fields = MakeFieldHolder([]string{
		"dgraph.type",
		"value",
	})
	Find(&v, rootUid, Fields)
	if len(v.Type) == 0 {
		t.FailNow()
	}
}

func TestHas(t *testing.T) {
	var v []Value
	err := FindHas(&v, "value", defaultFields)
	if len(v) == 0 {
		t.Log(err)
		t.Fail()
	}
	if v[0].Uid == "" {
		t.Fail()
	}
}

func TestFacetMutation(t *testing.T) {
	var v1 Value
	err := Find(&v1, rootUid, defaultFields)
	if err != nil {
		t.Fail()
	}
	m := FacetMutation(&v1, &v1.UidVal, "value3", "facetValue", "swagger")
	_, err = MutateMany(true, m)
	if err != nil {
		t.Fail()
	}
	err = Find(&v1, rootUid, defaultFields)
	if err != nil {
		t.Fail()
	}
	if v1.UidVal.FacetValue != "swagger" {
		t.Fail()
	}
}

func TestOnlyFields(t *testing.T) {
	q := NewQuery().SetField(defaultFields)
	err := q.Execute(nil)
	if err == nil {
		t.Fail()
	}
}

func TestFindByPredicate(t *testing.T) {
	var v Value
	err := FindByPredicate(&v, "value", TypeStr, defaultFields, "Test")
	if err != nil {
		t.Fail()
	}
	if v.StringVal != "Test" {
		t.Fail()
	}
}

//Append to array test, not singular
func TestAppendUid(t *testing.T) {
	var root, v1, v2, v3 Value
	Find(&root, rootUid, defaultFields)
	if len(root.UidArrayVal) == 0 {
		t.Fail()
	}
	v1 = root.UidArrayVal[5]
	v2 = root.UidArrayVal[6]
	v3 = root.UidArrayVal[7]
	_, err := ListAddUid(&v1, "value4", true, v2.Uid, v3.Uid)
	if err != nil {
		t.Fail()
	}
	u := v1.Uid
	v1 = Value{}
	Find(&v1, u, defaultFields)
	if len(v1.UidArrayVal) != 2 {
		t.Fail()
	}
}

func TestSetSingleUID(t *testing.T) {
	var v1 Value
	Find(&v1, rootUid, defaultFields)
	v2 := v1.UidArrayVal[2]
	v3 := v1.UidArrayVal[3]
	v4 := v1.UidArrayVal[4]
	u := v2.Uid
	err := SetSingleUID(&v2, "value3", v3.Uid, true)
	if err != nil {
		t.Fail()
	}
	v2 = Value{}
	Find(&v2, u, defaultFields)
	if v2.UidVal.Uid != v3.Uid {
		t.Fail()
	}
	if v3.Uid == "" {
		t.Fail()
	}
	u3 := v3.Uid
	err = SetSingleUID(&v2, "value3", v4.Uid, true)
	if err != nil {
		t.FailNow()
	}
	u4 := v4.Uid
	Find(&v3, u3, defaultFields)
	if v3.Uid != u3 {
		t.Log("Critical error, set single deletes more than relation.")
		t.FailNow()
	}
	if err != nil {
		t.Fail()
	}
	Find(&v2, u, defaultFields)
	if v2.UidVal.Uid != u4 {
		t.Fail()
	}
}

func TestPerformQueries(t *testing.T) {
	q1 := NewQuery().
		SetFunction(MakeFunction("eq").
			AddPredMultiple("value", TypeStr, "Test", "Test1", "Test2")).
		SetField(defaultFields)
	q2 := NewQuery().
		SetFunction(MakeFunction(FunctionHas).AddPred("value")).SetField(defaultFields)
	qu := q1.Append(q2)
	var v1, v2 []Value
	err := qu.Execute(&v1, &v2)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestUidToInt(t *testing.T) {
	str := "0x61FDA"
	val := UidToInt(str)
	if val != 401370 {
		t.Fail()
	}
}

func TestUidToIntString(t *testing.T) {
	str := "0x61FDA"
	val := UidToIntString(str)
	if val != "401370" {
		t.Fail()
	}
}

func TestSchemaFields(t *testing.T) {
	fi := getSchemaField("value")
	if fi.Type != TypeStr {
		t.Fail()
	}
}

func TestDeserialize(t *testing.T) {
	str := `{"uid":"","value":"Value","value2":1,"value4":null,"value5":null,"value6":null}`
	var v Value
	err := Deserialize(str, &v)
	if err != nil {
		t.Fail()
	}
	if v.IntVal != 1 || v.StringVal != "Value" {
		t.Fail()
	}
}

func TestDeleteUIDS(t *testing.T) {
	var v1, v2, v3 Value
	Find(&v1, "0x30", defaultFields)
	Find(&v2, "0x31", defaultFields)
	Find(&v3, "0x32", defaultFields)
	DeleteUIDS(true, "0x30", "0x31", "0x32")
	v1 = Value{}
	v2 = Value{}
	v3 = Value{}
	Find(&v1, "0x30", defaultFields)
	Find(&v2, "0x31", defaultFields)
	Find(&v3, "0x32", defaultFields)
	if v1.Uid != "" || v2.Uid != "" || v3.Uid != "" {
		t.Fail()
	}
}
