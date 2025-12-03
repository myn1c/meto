package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type SensorReading struct {
	Temp      float64 `json:"temp"`
	Humidity  float64 `json:"humidity"`
	Timestamp int64   `json:"timestamp"`
}

var (
	sensorHistory []SensorReading
	clients       = make(map[*websocket.Conn]bool)
	clientsMutex  = sync.Mutex{}
	upgrader      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func broadcast(reading SensorReading) {
	message, _ := json.Marshal(reading)

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Ошибка при отправке:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

func main() {
	r := gin.Default()
	
	r.GET("/", func(c *gin.Context) {
		lastTemp := "--"
		lastHum := "--"
		if len(sensorHistory) > 0 {
			last := sensorHistory[len(sensorHistory)-1]
			lastTemp = fmt.Sprintf("%.1f", last.Temp)
			lastHum = fmt.Sprintf("%.1f", last.Humidity)
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Температура и влажность</title>
<style>
body {
    font-family: 'Segoe UI', sans-serif;
    background: #f5f5f5;
    margin:0; height:100vh;
    display:flex; justify-content:center; align-items:center;
}
.card {
    background:#fff; padding:30px 40px;
    border-radius:12px;
    box-shadow:0 4px 12px rgba(0,0,0,0.1);
    width:300px; text-align:center;
}
h1 {color:#1e1e1e; margin-bottom:20px; font-size:22px;}
.tabs {display:flex; justify-content:space-around; margin-bottom:20px;}
.tab {flex:1; padding:10px 0; cursor:pointer; border-bottom:2px solid transparent; color:#555; font-weight:500;}
.tab.active {border-color:#4680c2; color:#4680c2;}
.value {font-size:48px; font-weight:bold; margin:20px 0;}
.temp {color:#e74c3c;}
.hum {color:#3498db;}
.status {margin-top:20px; color:#27ae60; font-weight:bold;}
</style>
</head>
<body>
<div class="card">
<h1>Мониторинг климата</h1>
<div class="tabs">
<div id="tab-temp" class="tab active">Температура</div>
<div id="tab-hum" class="tab">Влажность</div>
</div>
<div id="temp" class="value temp">%s °C</div>
<div id="hum" class="value hum" style="display:none;">%s %%</div>
<div id="status" class="status">Соединение...</div>
</div>
<script>
const tempEl = document.getElementById("temp");
const humEl = document.getElementById("hum");
const statusEl = document.getElementById("status");
const tabTemp = document.getElementById("tab-temp");
const tabHum = document.getElementById("tab-hum");

tabTemp.onclick = ()=>{
    tabTemp.classList.add("active");
    tabHum.classList.remove("active");
    tempEl.style.display="block";
    humEl.style.display="none";
};
tabHum.onclick = ()=>{
    tabHum.classList.add("active");
    tabTemp.classList.remove("active");
    humEl.style.display="block";
    tempEl.style.display="none";
};

const WS_URL = "ws://"+location.hostname+":8080/ws";
let ws = new WebSocket(WS_URL);

ws.onopen = ()=>{
    statusEl.textContent="Подключено ✓";
    statusEl.style.color="#27ae60";
};
ws.onmessage = (event)=>{
    const data=JSON.parse(event.data);
    tempEl.textContent=data.temp+" °C";
    humEl.textContent=data.humidity+" %";
};
ws.onclose = ()=>{
    statusEl.textContent="Нет соединения ×";
    statusEl.style.color="#e74c3c";
    setTimeout(()=>{ws=new WebSocket(WS_URL); ws.onopen=ws.onopen; ws.onmessage=ws.onmessage; ws.onclose=ws.onclose; ws.onerror=ws.onerror;},3000);
};
ws.onerror = ()=>{
    statusEl.textContent="Ошибка соединения";
    statusEl.style.color="#e74c3c";
};
</script>
</body>
</html>`, lastTemp, lastHum)

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})
	
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Ошибка апгрейда:", err)
			return
		}

		clientsMutex.Lock()
		clients[conn] = true
		clientsMutex.Unlock()
		
		if len(sensorHistory) > 0 {
			last := sensorHistory[len(sensorHistory)-1]
			conn.WriteJSON(last)
		}

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				clientsMutex.Lock()
				delete(clients, conn)
				clientsMutex.Unlock()
				conn.Close()
				break
			}
		}
	})
	
	r.GET("/data", func(c *gin.Context) {
		tempStr := c.Query("temp")
		humStr := c.Query("humidity")

		var temp, hum float64
		_, err1 := fmt.Sscan(tempStr, &temp)
		_, err2 := fmt.Sscan(humStr, &hum)
		if err1 != nil || err2 != nil {
			c.JSON(400, gin.H{"error": "invalid parameters"})
			return
		}

		reading := SensorReading{
			Temp:      temp,
			Humidity:  hum,
			Timestamp: time.Now().Unix(),
		}
		
		sensorHistory = append(sensorHistory, reading)
		if len(sensorHistory) > 1000 {
			sensorHistory = sensorHistory[1:]
		}
		
		broadcast(reading)

		c.JSON(200, gin.H{"status": "ok", "received": reading})
	})

	r.Run(":8080")
}
