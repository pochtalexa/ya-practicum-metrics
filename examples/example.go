package main

import (
	"fmt"
	// "time"
	_ "encoding/json"
	"net/http"	
	_ "reflect"
	_ "runtime"
	_ "strconv"
	"io"
	"log"
)

// var rtm runtime.MemStats


func MyMain() {
	
	/*
	runtime.ReadMemStats(&rtm)
	
	// b, _ := json.Marshal(rtm)
	// fmt.Println(b)
	// fmt.Println(string(b))
	fmt.Println(rtm.Alloc)

	fields := reflect.TypeOf(rtm)
	values := reflect.ValueOf(rtm)
	num := fields.NumField()

	for i := 0; i < num; i++ {
    	field := fields.Field(i)
    	value := values.Field(i)
    	fmt.Println("Type:", field.Type, ",", field.Name, "=", value)

		// switch value.Kind() {
		// case reflect.String:
		// 	v := value.String()
		// 	fmt.Println(v)
		// case reflect.Int:
		// 	v := strconv.FormatInt(value.Int(), 10)
		// 	fmt.Println(v)
		// case reflect.Int32:
		// 	v := strconv.FormatInt(value.Int(), 10)
		// 	fmt.Println(v)
		// case reflect.Int64:
		// 	v := strconv.FormatInt(value.Int(), 10)
		// 	fmt.Println(v)
		// default:
			// assert.Fail(t, "Not support type of struct")
		// }
	}
	*/

	req, err := http.NewRequest("GET", "https://icanhazdadjoke.com", nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Accept", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)	
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(b))


	/*
	res, err := http.Get("https://ya.ru")

	b, err = io.ReadAll(res.Body)	
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(b))
	*/

}
