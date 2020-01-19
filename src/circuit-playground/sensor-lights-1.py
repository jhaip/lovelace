import time
import board
import neopixel
import busio
from adafruit_circuitplayground.express import cpx

N_PIXELS = 10
pixels = neopixel.NeoPixel(board.NEOPIXEL, 10, brightness=0.2, auto_write=False)
uart = busio.UART(board.TX, board.RX, baudrate=115200)
last_sensor_reading = time.monotonic()

button_a = digitalio.DigitalInOut(board.BUTTON_A)
button_a.switch_to_input(pull=digitalio.Pull.DOWN)
button_b = digitalio.DigitalInOut(board.BUTTON_A)
button_b.switch_to_input(pull=digitalio.Pull.DOWN)
light = analogio.AnalogIn(board.LIGHT)

while True:
    # Read commands to change outputs
    data = uart.readline()
    if data is not None:
        parsed_data = data.rstrip().split(b",")
        if len(parsed_data) > 1:
            prefix = parsed_data[0]
            if prefix == b'LIGHT' and len(parsed_data) == 5:
                print("PARSING LIGHT COMMAND: {}".format(data))
                pixel_id = int(parsed_data[1])
                if pixel_id >= 0 and parsed_data < N_PIXELS:
                    r = int(parsed_data[2])
                    g = int(parsed_data[3])
                    b = int(parsed_data[4])
                    pixels[pixel_id] = (r, g, b)
                    time.sleep(0.2)
                    pixels.show()
    
    # Write sensor values
    if time.monotonic() - last_sensor_reading > 1000:
        last_sensor_reading = time.monotonic()
        button_a_value = 1 if button_a.value else 0
        button_b_value = 1 if button_b.value else 0
        light_value = light.value
        print("SENDING SENSOR VALUES: {} {} {}".format(button_a_value, button_b_value, light_value))
        uart.write(b"BUTTON_A:{}\n".format(button_a_value))
        uart.write(b"BUTTON_B:{}\n".format(button_b_value))
        uart.write(b"LIGHT:{}\n".format(light_value))

    time.sleep(0.1)
    