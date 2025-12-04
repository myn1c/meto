package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Это автоматический настройщик клиентской части meto")

	SSID := prompt(reader, "Введите название сети ESP: ")
	PASS := prompt(reader, "Введите пароль от сети: ")
	PORT := promptDefault(reader, "Введите порт (по умолчанию 8080): ", "8080")
	HOST := prompt(reader, "Введите IP host (или Auto для локального IP): ")
	if strings.ToLower(HOST) == "auto" || HOST == "" {
		HOST = getLocalIP()
	}

	fmt.Println("\nНазвание сети:", SSID)
	fmt.Println("Пароль:", PASS)
	fmt.Println("Порт:", PORT)
	fmt.Println("Хост:", HOST)

	clientFile := "client_esp32.py"
	if err := createClientFile(clientFile, SSID, PASS, HOST, PORT); err != nil {
		fmt.Println("Не удалось создать", clientFile, ":", err)
		return
	}
	fmt.Println("Файл", clientFile, "успешно создан!")

	fmt.Println("\nПолучаем список релизов сервера с GitHub...")
	releases := getGitHubReleases()
	serverReleases := filterServerReleases(releases)
	if len(serverReleases) == 0 {
		fmt.Println("Релизов с сервером не найдено")
		return
	}

	fmt.Println("Найденные серверные релизы:")
	for i, r := range serverReleases {
		fmt.Printf("[%d] %s\n", i+1, r.TagName)
	}

	choice := prompt(reader, "Выберите номер релиза для скачивания: ")
	index := 0
	fmt.Sscanf(choice, "%d", &index)
	if index < 1 || index > len(serverReleases) {
		fmt.Println("Некорректный выбор")
		return
	}

	selected := serverReleases[index-1]
	var assetIndex int = -1
	for i, a := range selected.Assets {
		if strings.Contains(strings.ToLower(a.Name), "server") {
			assetIndex = i
			break
		}
	}
	if assetIndex == -1 {
		fmt.Println("Не найден asset с 'server' в имени у выбранного релиза")
		return
	}
	asset := selected.Assets[assetIndex]
	serverFile := asset.Name
	fmt.Println("Скачиваем", serverFile)
	downloadFile(serverFile, asset.BrowserDownloadURL)
	fmt.Println("Сервер сохранён:", serverFile)

	if _, err := os.Stat(clientFile); os.IsNotExist(err) {
		fmt.Println(clientFile, "отсутствует - создаём заново перед запуском сервера...")
		if err := createClientFile(clientFile, SSID, PASS, HOST, PORT); err != nil {
			fmt.Println("Не удалось создать", clientFile, ":", err)
			return
		}
		fmt.Println("Создание", clientFile, "завершено.")
	}

	run := prompt(reader, "Запустить сервер? (y/n): ")
	if strings.ToLower(run) == "y" {
		fmt.Println("Убедились, что", clientFile, "на месте - запускаем", serverFile)
		cmd := exec.Command("./" + serverFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Ошибка при запуске сервера:", err)
		}
	}
}

func prompt(reader *bufio.Reader, text string) string {
	for {
		fmt.Print(text)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			return input
		}
	}
}

func promptDefault(reader *bufio.Reader, text string, def string) string {
	fmt.Print(text)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func createClientFile(path, ssid, pass, host, port string) error {
	content := fmt.Sprintf(`import network
import usocket as socket
import time
import ujson
from machine import Pin
import dht

SSID = "%s"
PASS = "%s"

API_HOST = "%s"
API_PORT = %s

wifi = network.WLAN(network.STA_IF)
wifi.active(True)
wifi.connect(SSID, PASS)

print("Подключение к Wi-Fi...")
while not wifi.isconnected():
    time.sleep(0.1)

print("Wi-Fi подключён:", wifi.ifconfig()[0])

sensor = dht.DHT11(Pin(4))

def send_data(temp, humidity):
    try:
        addr = socket.getaddrinfo(API_HOST, API_PORT)[0][-1]
        s = socket.socket()
        s.settimeout(5)
        s.connect(addr)

        req = "GET /data/?temp={}&humidity={} HTTP/1.1\\r\\n".format(temp, humidity)
        req += "Host: {}\\r\\n".format(API_HOST)
        req += "Connection: close\\r\\n\\r\\n"
        s.send(req.encode())

        data = s.recv(1024)
        print("Ответ сервера:", data.decode())
    except Exception as e:
        print("Ошибка:", e)
    finally:
        s.close()

while True:
    try:
        sensor.measure()
        temp = sensor.temperature()
        humidity = sensor.humidity()
        print("Температура:", temp, "Влажность:", humidity)
        send_data(temp, humidity)
    except Exception as e:
        print("Ошибка с датчиком:", e)
    time.sleep(5)
`, ssid, pass, host, port)

	return os.WriteFile(path, []byte(content), 0644)
}

func getGitHubReleases() []Release {
	url := "https://api.github.com/repos/myn1c/meto/releases"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка запроса к GitHub:", err)
		return nil
	}
	defer resp.Body.Close()

	var releases []Release
	json.NewDecoder(resp.Body).Decode(&releases)
	return releases
}

func filterServerReleases(releases []Release) []Release {
	var filtered []Release
	for _, r := range releases {
		for _, a := range r.Assets {
			if strings.Contains(strings.ToLower(a.Name), "server") {
				filtered = append(filtered, r)
				break
			}
		}
	}
	return filtered
}

func downloadFile(filePath, url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка скачивания:", err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer out.Close()

	io.Copy(out, resp.Body)
}
