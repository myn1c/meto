<p align="center">
    <img src="doc/meto.png" width="100" height="100">
    <br>
    <b>METO</b>
    <br>
</p>


### Чтобы использовать данный проект, используйте автоматический установщик.

Пример работы с installer.go
```
Это автоматический настройщик клиентской части meto
Введите название сети ESP: Ufanet
Введите пароль от сети: Password
Введите порт (по умолчанию 8080): 
Введите IP host (или Auto для локального IP): Auto

Название сети: Ufanet
Пароль: Password
Порт: 8080
Хост: 169.254.142.85
Файл client_esp32.py успешно создан!

Получаем список релизов сервера с GitHub...
Найденные серверные релизы:
[1] windows
[2] kubuntu
Выберите номер релиза для скачивания: 1
Скачиваем server.exe
Сервер сохранён: server.exe
Запустить сервер? (y/n): y
Убедились, что client_esp32.py на месте - запускаем server.exe
```

### Вручную
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


## Пример консоли работы сервера
![cmd](doc/Cmd.png)
