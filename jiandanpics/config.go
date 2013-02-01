package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type config struct {
	file string
}

func (m *config) Load(v interface{}) error {
	jsonBlob, err := ioutil.ReadFile(m.file)
	if err != nil {
		log.Println("ioutil.ReadFile:", err)
		return err
	}

	err = json.Unmarshal(jsonBlob, v)
	if err != nil {
		log.Println("json.Unmarshal:", err)
		return err
	}

	return nil
}

func (m *config) Save(v interface{}) error {
	f, err := os.Create(m.file)
	if err != nil {
		log.Println("os.Create:", err)
		return err
	}
	defer f.Close()

	b, err := json.Marshal(v)
	if err != nil {
		log.Println("json.Marshal:", err)
		return err
	}

	_, err = f.Write(b)
	if err != nil {
		log.Println("file.Write:", err)
		return err
	}

	return nil
}
