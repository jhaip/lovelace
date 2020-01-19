import time
import board
import neopixel
import busio
import digitalio
uart = busio.UART(board.TX, board.RX, baudrate=115200, timeout=1)
from adafruit_circuitplayground import cp

N_PIXELS = 10
last_sensor_reading = time.monotonic()

while True:
    # Read commands to change outputs
    data = uart.readline()
    if data is not None:
        print(data)
        parsed_data = data.rstrip().split(b",")
        if len(parsed_data) > 1:
            prefix = parsed_data[0]
            if prefix == b'LIGHT' and len(parsed_data) == 5:
                print("PARSING LIGHT COMMAND: {}".format(data))
                pixel_id = int(parsed_data[1])
                if pixel_id >= 0 and pixel_id < N_PIXELS:
                    r = int(parsed_data[2])
                    g = int(parsed_data[3])
                    b = int(parsed_data[4])
                    cp.pixels[pixel_id] = (r, g, b)
    else:
        print("DATA WAS NONE")

    # Write sensor values
    if time.monotonic() - last_sensor_reading > 0.1:
        last_sensor_reading = time.monotonic()
        button_a_value = 1 if cp.button_a else 0
        button_b_value = 1 if cp.button_b else 0
        light_value = cp.light
        print("SENDING SENSOR VALUES: {} {} {}".format(button_a_value, button_b_value, light_value))
        uart.write(b"BUTTON_A:{}\n".format(button_a_value).encode())
        uart.write(b"BUTTON_B:{}\n".format(button_b_value).encode())
        uart.write(b"LIGHT:{}\n".format(light_value).encode())

    time.sleep(0.1)