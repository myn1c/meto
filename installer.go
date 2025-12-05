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

type ServerRelease struct {
	OSName string
	Asset  struct {
		Name string
		URL  string
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Это автоматический настройщик клиентской части meto")
	SSID, PASS, PORT, HOST := input()

	clientURL := "https://raw.githubusercontent.com/myn1c/meto/main/src/client_esp32.py"
	clientFile := "client_esp32.py"

	fmt.Println("Скачиваем клиентский файл:", clientURL)
	origContent, err := httpGetText(clientURL)
	if err != nil {
		fmt.Println("Не удалось скачать client_esp32.py:", err)
		return
	}

	finalContent := buildClientContent(origContent, SSID, PASS, HOST, PORT)

	if err := os.WriteFile(clientFile, []byte(finalContent), 0644); err != nil {
		fmt.Println("Ошибка записи файла:", err)
		return
	}
	fmt.Println("Файл", clientFile, "успешно создан!")
	run := prompt(reader, "Скачать сервер? (y/n): ")
	if strings.ToLower(run) == "y" {
		fmt.Println("\nПолучаем список релизов сервера с GitHub...")
		releases := getGitHubReleases()
		servers := findServerReleases(releases)
		if len(servers) == 0 {
			fmt.Println("Релизов с сервером не найдено")
			return
		}

		fmt.Println("Найденные серверные релизы:")
		for i, s := range servers {
			fmt.Printf("[%d] %s\n", i+1, s.OSName)
		}

		choice := prompt(reader, "Выберите номер релиза для скачивания: ")
		index := 0
		fmt.Sscanf(choice, "%d", &index)
		if index < 1 || index > len(servers) {
			fmt.Println("Некорректный выбор")
			return
		}

		selected := servers[index-1]
		serverFile := selected.Asset.Name
		fmt.Println("Скачиваем", serverFile)
		downloadFile(serverFile, selected.Asset.URL)

		run := prompt(reader, "Запустить сервер? (y/n): ")
		if strings.ToLower(run) == "y" {
			cmd := exec.Command("./" + serverFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("Ошибка при запуске сервера:", err)
			}
		}
	}

	fmt.Println("Инсталлер завершил свою работу")
}

func input() (string, string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	SSID := prompt(reader, "Введите название сети ESP: ")
	PASS := prompt(reader, "Введите пароль от сети: ")
	PORT := promptDefault(reader, "Введите порт (по умолчанию 8080): ", "8080")
	HOST := promptDefault(reader, "Введите IP host (по умолчанию локальный IP пк в сети): ", getLocalIP())

	for {
		fmt.Println("\nТекущие настройки:")
		fmt.Println("1. Название сети (SSID):", SSID)
		fmt.Println("2. Пароль (PASS):", PASS)
		fmt.Println("3. Порт (PORT):", PORT)
		fmt.Println("4. Хост (HOST):", HOST)
		fmt.Println("5. Всё верно, продолжить")

		choice := prompt(reader, "Введите номер переменной, которую хотите изменить (или 5 для продолжения): ")

		switch choice {
		case "1":
			SSID = prompt(reader, "Введите новое название сети ESP: ")
		case "2":
			PASS = prompt(reader, "Введите новый пароль: ")
		case "3":
			PORT = promptDefault(reader, "Введите новый порт: ", PORT)
		case "4":
			HOST = promptDefault(reader, "Введите новый хост: ", HOST)
		case "5":
			return SSID, PASS, PORT, HOST
		default:
			fmt.Println("Некорректный выбор, попробуйте ещё раз.")
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

func promptDefault(reader *bufio.Reader, text, def string) string {
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
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}

func httpGetText(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ошибка %d: файл не найден по URL %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	return string(data), err
}

func buildClientContent(orig, ssid, pass, host, port string) string {
	header := fmt.Sprintf(`# --- Auto generated config ---
SSID = "%s"
PASS = "%s"
API_HOST = "%s"
API_PORT = %s

`, ssid, pass, host, port)

	return header + orig
}

func getGitHubReleases() []Release {
	url := "https://api.github.com/repos/myn1c/meto/releases"
	for {
		token := os.Getenv("GITHUB_TOKEN")
		releases, err := fetchReleases(url, token)
		if err == nil {
			return releases
		}

		if strings.Contains(err.Error(), "API rate limit exceeded") {
			fmt.Println("Превышен лимит запросов к GitHub.")
			fmt.Println("Чтобы продолжить, нужно ввести Personal Access Token (PAT).")
			fmt.Println("Инструкция:")
			fmt.Println("1. Зайди на https://github.com/settings/tokens")
			fmt.Println("2. Developer settings → Personal access tokens → Tokens (classic) → Generate new token")
			fmt.Println("3. Для чтения релизов публичных репозиториев достаточно любого токена")
			fmt.Print("Введите токен: ")

			reader := bufio.NewReader(os.Stdin)
			inputToken, _ := reader.ReadString('\n')
			inputToken = strings.TrimSpace(inputToken)
			token = inputToken

			releases, err := fetchReleases(url, token)
			if err != nil {
				fmt.Println("Ошибка запроса с токеном:", err)
				return nil
			}
			return releases
		} else {
			fmt.Println("Ошибка запроса GitHub:", err)
			return nil
		}
	}
}

func fetchReleases(url, token string) ([]Release, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(string(body))
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func findServerReleases(releases []Release) []ServerRelease {
	var out []ServerRelease
	for _, r := range releases {
		for _, a := range r.Assets {
			lower := strings.ToLower(a.Name)
			if strings.Contains(lower, "server") {
				idx := strings.Index(lower, "server")
				rest := ""
				if idx != -1 {
					rest = lower[idx+len("server"):]
				}

				rest = strings.TrimLeft(rest, "-_.")
				if dot := strings.Index(rest, "."); dot != -1 {
					rest = rest[:dot]
				}
				if rest == "" {
					rest = r.TagName
				}
				sr := ServerRelease{
					OSName: rest,
				}
				sr.Asset.Name = a.Name
				sr.Asset.URL = a.BrowserDownloadURL
				out = append(out, sr)
				break
			}
		}
	}
	return out
}

func downloadFile(path, url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка загрузки:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Ошибка: сервер вернул %d при попытке скачать %s\n", resp.StatusCode, url)
		return
	}

	out, err := os.Create(path)
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		fmt.Println("Ошибка записи в файл:", err)
	} else {
		fmt.Println("Сервер сохранён:", path)
	}
}
