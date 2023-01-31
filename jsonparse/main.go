package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type DataAContents struct {
	Prop string `json:"propForTypeA"`
}

type DataAMisc struct {
	Prop1 string `json:"prop1"`
	Prop2 string `json:"prop2"`
}

type DataA struct {
	Contents DataAContents `json:"contents"`
	Misc     DataAMisc     `json:"miscForA"`
}

type DataBContents struct {
	Prop string `json:"propForTypeB"`
}

type DataBMisc struct {
	Prop1 string `json:"prop1"`
	Prop2 string `json:"prop2"`
}

type DataB struct {
	Contents DataBContents `json:"contents"`
	Misc     DataBMisc     `json:"miscForB"`
}

// DataAny marshalls/unmarshalls from/to DataA or DataB.
type DataAny struct {
	// value is DatA or DataB.
	value any
}

func (d *DataAny) MarshalJSON() ([]byte, error) {

	switch t := d.value.(type) {
	case *DataA:
		type DataAJSON struct {
			Type string `json:"type"`
			DataA
		}
		return json.Marshal(DataAJSON{"a", *t})
	case *DataB:
		type DataBJSON struct {
			Type string `json:"type"`
			DataB
		}
		return json.Marshal(DataBJSON{"b", *t})
	default:
		return []byte(""), errors.New("unknown value type")
	}
}

func (d *DataAny) UnmarshalJSON(data []byte) error {
	var objmap map[string]json.RawMessage
	err := json.Unmarshal(data, &objmap)
	if err != nil {
		return err
	}

	var objType string
	err = json.Unmarshal(objmap["type"], &objType)

	switch objType {
	case "a":
		var value DataA
		err = json.Unmarshal(data, &value)
		if err != nil {
			return err
		}
		d.value = &value
	case "b":
		var value DataB
		err = json.Unmarshal(data, &value)
		if err != nil {
			return err
		}
		d.value = &value
	case "":
		return errors.New("no object type set")
	default:
		return fmt.Errorf("unknown object type %s", objType)
	}

	return nil
}

func exit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	fileA := "assets/example_a.json"
	fileB := "assets/example_b.json"

	exampleAJson, err := ioutil.ReadFile(fileA)
	if err != nil {
		exit(err)
	}
	exampleBJson, err := ioutil.ReadFile(fileB)
	if err != nil {
		exit(err)
	}

	var dataA DataA
	err = json.Unmarshal(exampleAJson, &dataA)
	if err != nil {
		exit(err)
	}

	dataAJson, err := json.Marshal(&dataA)
	if err != nil {
		exit(err)
	}

	fmt.Printf("JSON from %s parsed->serialized as DataA: %s\n", fileA, dataAJson)

	var dataB DataB
	err = json.Unmarshal(exampleBJson, &dataB)
	if err != nil {
		exit(err)
	}

	dataBJson, err := json.Marshal(&dataB)
	if err != nil {
		exit(err)
	}

	fmt.Printf("JSON from %s parsed->serialized as DataB: %s\n", fileB, dataBJson)

	var dataAny DataAny
	err = json.Unmarshal(exampleAJson, &dataAny)
	if err != nil {
		exit(err)
	}

	dataAnyJson, err := json.Marshal(&dataAny)
	if err != nil {
		exit(err)
	}

	fmt.Printf("JSON from %s parsed-serialized as DataAny: %s\n", fileB, dataAnyJson)
}
