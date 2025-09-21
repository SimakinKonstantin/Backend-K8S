package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

// Тесты добавления.
func TestAddingHundred(t *testing.T) {

	// Получаем первоначальное кол-во элементов
	programs, _ := http.Get("http://localhost:8081/programs/count")
	defer programs.Body.Close()
	body, _ := ioutil.ReadAll(programs.Body)
	var programsCount int64
	json.Unmarshal(body, &programsCount)

	// Добавление 100 элементов.
	for i := 0; i < 100; i++ {
		http.Post("http://localhost:8081/programs/", "application/x-www-form-urlencoded",
			bytes.NewBufferString("Name=1&Description=1&Price=1&ConfirmUserId=6739bebe108bb3da33766377"))
	}

	programs, _ = http.Get("http://localhost:8081/programs/count")
	defer programs.Body.Close()
	body, _ = ioutil.ReadAll(programs.Body)
	var programsCount2 int64
	json.Unmarshal(body, &programsCount2)

	assert.Equal(t, int64(programsCount)+100, programsCount2,
		"До добавления + 100 должно быть == после добавления")
}

func TestAddingHundredThousand(t *testing.T) {

	// Получаем первоначальное кол-во элементов
	programs, _ := http.Get("http://localhost:8081/programs/count")
	defer programs.Body.Close()
	body, _ := ioutil.ReadAll(programs.Body)
	var programsCount int64
	json.Unmarshal(body, &programsCount)

	// Добавление 100 000 элементов.
	for i := 0; i < 100000; i++ {
		http.Post("http://localhost:8081/programs/",
			"application/x-www-form-urlencoded", bytes.NewBufferString("Name=1&Description=1&Price=1"))
	}

	programs, _ = http.Get("http://localhost:8081/programs/count")
	defer programs.Body.Close()
	body, _ = ioutil.ReadAll(programs.Body)
	var programsCount2 int64
	json.Unmarshal(body, &programsCount2)

	assert.Equal(t, int64(programsCount)+100000, int64(programsCount2),
		"До добавления + 100 000 должно быть == после добавления")
}

func TestDeleting(t *testing.T) {
	programs, _ := http.Get("http://localhost:8081/programs/count")
	defer programs.Body.Close()
	body, _ := ioutil.ReadAll(programs.Body)
	var programsCount int64
	json.Unmarshal(body, &programsCount)

	client := http.Client{}
	req, _ := http.NewRequest("DELETE", "http://localhost:8081/programs", nil)
	client.Do(req)

	programs, _ = http.Get("http://localhost:8081/programs/count")
	defer programs.Body.Close()
	body, _ = ioutil.ReadAll(programs.Body)
	var programsCount2 int64
	json.Unmarshal(body, &programsCount2)

	assert.Equal(t, int64(0), int64(programsCount2), "После удаления додлжно остаться 0")
}
