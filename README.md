# meto
### Чтобы использовать данный сервер

Замените переменные в файле client_esp32.py
```
SSID = "Название сети"
PASS = "Пароль сети"

API_HOST = "ip хоста"
API_PORT = порт
```

ip хоста можно узнать с помощью данного кода 
```python
import socket

def get_local_ip():
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    try:
        s.connect(("8.8.8.8", 80))
        return s.getsockname()[0]
    except:
        return None
    finally:
        s.close()

print("IP:", get_local_ip())
```

Команда для запуска сервера на linux
```bash
./server
```


### Сайт выглядеть так по пути /
![Скрин сайта](doc/Metoсайта.jpeg)

## Эндпоинты
```
../ - возвращает страницу
../data/ - нужен для добавления изменений (temp="температура", humidity="влажность")
```
