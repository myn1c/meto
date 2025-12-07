import network
import usocket as socket
import time
import ujson
from machine import Pin, ADC
import dht


# SSID = "%s"
# PASS = "%s"

# API_HOST = "%s"
# API_PORT = %s


wifi = network.WLAN(network.STA_IF)
wifi.active(True)
wifi.connect(SSID, PASS)

print("Подключение к Wi-Fi...")
while not wifi.isconnected():
    time.sleep(0.1)

print("Wi-Fi подключён:", wifi.ifconfig()[0])

sensor = dht.DHT11(Pin(4))
adc = ADC(Pin(36))
adc.atten(ADC.ATTN_11DB)

def read_internal_temp():
    raw = adc.read()
    voltage = raw / 4095 * 3.3
    temp_c = 27 - (voltage - 0.706)/0.001721
    return round(temp_c, 1)
    
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
        try:
            sensor.measure()
            temp = sensor.temperature()
            humidity = sensor.humidity()
        except Exception as e:
            print("DHT11 не доступен, используем встроенный датчик:", e)
            temp = read_internal_temp()
            humidity = 0

        print("Температура:", temp, "°C", "Влажность:", humidity)
        send_data(temp, humidity)

    except Exception as e:
        print("Ошибка основного цикла:", e)
    time.sleep(5)
