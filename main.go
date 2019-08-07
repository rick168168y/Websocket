package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	//Don't forgot to add import _ "image/jpeg" the image package itself doesn't know how to decode jpeg, you need to import image/jpeg to register the jpeg decoder.
)

type Message struct {
	//訊息struct
	Operation string `json:"operation,omitempty"` //做什麼
	Content   string `json:"content,omitempty"`   //內容
	Filename  string `json:"filename,omitempty"`
	Chunks    int    `json:"chunks,omitempty"`  //分成幾次上傳
	Partial   int    `json:"partial,omitempty"` //上傳的第幾個（only for files）
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  10000000,
	WriteBufferSize: 10000000,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		var m Message
		err = json.Unmarshal(message, &m)
		fmt.Println(m.Operation)
		if err != nil {
			fmt.Println("error:", err)
		} else {
			switch m.Operation {
			case "UploadRequest":
				response, _ := json.Marshal(&Message{Operation: "UploadResponse"})
				conn.WriteMessage(websocket.TextMessage, response)
				filename := m.Filename
				chunks := m.Chunks
				var index int
				var filecontent []byte
				//var filecontent string
				index = 0
				fmt.Println(chunks)
				t1 := time.Now()
				for index < chunks {
					_, received, err := conn.ReadMessage()
					if err != nil {
						fmt.Println("Read Error:", err)
						break
					}
					err = json.Unmarshal(received, &m)
					//fmt.Println("len=", len(m.Content))
					//fmt.Println("count=", m.Partial)
					//filecontent = append(filecontent, m.Content...)
					decoded, _ := base64.StdEncoding.DecodeString(m.Content)
					filecontent = append(filecontent, decoded...)
					index++
				}
				var extension = filepath.Ext(string(filename))
				var name = filename[0 : len(filename)-len(extension)]
				if err != nil {
					break
				}
				//decoded, _ := base64.StdEncoding.DecodeString(filecontent)
				//fmt.Println("len=", len(filecontent))
				//fmt.Println("dec_len=", len(decoded))
				//ioutil.WriteFile("Fileupload/"+string(name)+strconv.FormatInt(time.Now().UnixNano(), 10)+"."+extension, filecontent, 0644)
				ioutil.WriteFile("Fileupload/"+string(name)+strconv.FormatInt(time.Now().UnixNano(), 10)+extension, filecontent, 0644)
				t2 := time.Now()
				diff := t2.Sub(t1)
				fmt.Println("Time Duration:", diff)
				//success := []byte("Upload successfully\n")
				//conn.WriteMessage(2, success)
				fmt.Println("Upload Successfully")
			}
		}

		/*_, filename, err := conn.ReadMessage()
		_, chunks, err := conn.ReadMessage()
		var index int
		index = 0
		var filecontent []byte
		intchunk, _ := strconv.Atoi(string(chunks))
		fmt.Println(intchunk)
		for index <=  intchunk{
			_, fileseg, _ := conn.ReadMessage()
			filecontent = append(filecontent, fileseg...)
			index++
		}

		var extension = filepath.Ext(string(filename))
		var name = filename[0 : len(filename)-len(extension)]
		if err != nil {
			break
		}
		ioutil.WriteFile("Fileupload/"+string(name)+strconv.FormatInt(time.Now().UnixNano(), 10)+"."+extension, filecontent, 0644)
		success := []byte("Upload successfully\n")
		conn.WriteMessage(2, success)*/
		//encodemessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
	}
}

func main() {
	r := gin.Default()
	r.GET("/websocket", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})
	r.Run() // listen and serve on 0.0.0.0:8080

}
