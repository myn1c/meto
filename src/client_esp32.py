import network
import usocket as socket
import time
import ujson
from machine import Pin
import dht

SSID = "Название сети"
PASS = "Пароль сети"

API_HOST = "ip хоста"
API_PORT = 'порт хоста !!! кавычки убрать'

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

        req = "GET /data/?temp={}&humidity={} HTTP/1.1\r\n".format(temp, humidity)
        req += "Host: {}\r\n".format(API_HOST)
        req += "Connection: close\r\n\r\n"
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
