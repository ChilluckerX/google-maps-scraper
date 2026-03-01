package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	b, err := os.ReadFile("d:/google-maps-scraper/empty_operation_hours.txt")
	if err != nil {
		fmt.Println("Error reading file", err)
		return
	}
	s := string(b)
	lines := strings.Split(s, "\n")
	var jsonStr string
	for _, l := range lines {
		if strings.HasPrefix(l, "    \"d\": \")]}'\\n") {
			jsonStr = strings.TrimPrefix(l, "    \"d\": \")]}'\\n")
			jsonStr = strings.TrimSuffix(jsonStr, "\",")
			jsonStr = strings.TrimSuffix(jsonStr, "\"")
			break
		}
	}

	// Need to unescape json string
	jsonStr = strings.ReplaceAll(jsonStr, `\n`, "\n")
	jsonStr = strings.ReplaceAll(jsonStr, `\"`, "\"")
	jsonStr = strings.ReplaceAll(jsonStr, `\\`, `\`)

	var dArr []any
	err = json.Unmarshal([]byte(jsonStr), &dArr)
	if err != nil {
		fmt.Println("json parse err", err)
	} else {
		fmt.Println("Success parsing!")
		d := dArr[0].([]any)[1].([]any)
		// check index 203
		if len(d) > 203 && d[203] != nil {
			fmt.Println("203:", len(d[203].([]any)))
		} else {
			fmt.Println("203 is nil or missing")
		}
		// check index 34
		if len(d) > 34 && d[34] != nil {
			fmt.Println("34:", len(d[34].([]any)))
		} else {
			fmt.Println("34 is nil or missing")
		}
	}
}
