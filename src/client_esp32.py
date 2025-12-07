
from machine import Pin
import dht

#SSID = "%s"
#PASS = "%s"

#API_HOST = "%s"
#API_PORT = %s

wifi = network.WLAN(network.STA_IF)
wifi.active(True)

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
