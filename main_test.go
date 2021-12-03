package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

//func Test_sort(t *testing.T) {
//
//}

func Test_getDepends(t *testing.T) {
	_, deps, err := getDepends()
	if err != nil {
		t.Error(err)
		return
	}
	for k, v := range deps {
		fmt.Println(k, v)
	}
}

func Test_time(t *testing.T) {
	fmt.Println(time.Now().Format(time.RFC3339Nano))
	fmt.Println(time.Now().UnixNano())

	time.Now().Format("2006-01-02 15:04")

	fmt.Println(time.Unix(0, 1637583921438266000).Format(time.RFC3339Nano))

	fmt.Println(time.Now().AddDate(0, 0, -3).UnixNano())
	fmt.Println(time.Now().UnixNano())
}

func Test_json(t *testing.T) {
	req := struct {
		Add []struct {
			Account          string    `json:"account"`
			Exchange         string    `json:"exchange"`
			ChainAccountType string    `json:"chain_account_type"`
			StartDate        time.Time `json:"start_date"`
		} `json:"add"`
		Delete []string `json:"delete"`
		Modify []struct {
			Name             string    `json:"name"`
			Account          string    `json:"account"`
			Exchange         string    `json:"exchange"`
			ChainAccountType string    `json:"chain_account_type"`
			StartDate        time.Time `json:"start_date"`
		} `json:"modify"`
	}{}

	str := `{"add":[
    {
        "account":"test1",
        "exchange":"ex1",
        "chain_account_type":"spot",
        "start_date":1637583921438266000
    }
], 
"delete": ["111", "222"],
"modify": [
    {}
]}`

	if err := json.Unmarshal([]byte(str), &req); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(req)
}

func Test_json2(t *testing.T) {
	str := `[
    {
        "name": "otchedge/ex1-c4",
        "account":"test001",
        "exchange":"ex001",
        "chain_account_type":"spot",
        "start_date":1637583921438266000
    },
    {
        "name": "test001",
    },
    {
        "account":"test003",
        "exchange":"ex003",
        "chain_account_type":"spot",
        "start_date":1637583921438266000
    }
]`

	var req []struct {
		Name             string `json:"name"`
		Account          string `json:"account"`
		Exchange         string `json:"exchange"`
		ChainAccountType string `json:"chain_account_type"`
		StartDate        int64  `json:"start_date"`
	}

	if err := json.Unmarshal([]byte(str), &req); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(req)
}

func Test_simpleCombine(t *testing.T) {
	_, data, err := getDepends()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 生成文件
	buf := &bytes.Buffer{}
	for k := range data {
		if k != "user" {
			continue
		}
		buf.WriteString("digraph G {\n")
		simpleCombine(buf, k, data)
		buf.WriteString("}")
		break
	}
}

func Test_isRing(t *testing.T) {
	// isRing()
}

func Test_Map(t *testing.T) {
	m := make(map[int]struct{})
	mapCap(m)
	fmt.Println(len(m))
}

func mapCap(m map[int]struct{}) {
	m[-1] = struct{}{}
	m[-2] = struct{}{}
	m[-3] = struct{}{}

	for i := 0; i < 1000000; i++ {
		m[i] = struct{}{}
	}
}
