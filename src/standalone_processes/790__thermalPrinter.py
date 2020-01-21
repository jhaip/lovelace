from helper2 import init, claim, retract, prehook, subscription, batch, get_my_id_str
import helper2
import logging
import board
import busio
import serial
import adafruit_thermal_printer

helper2.rpc_url = "10.0.0.22"

ThermalPrinter = adafruit_thermal_printer.get_printer_class(2.64)
uart = serial.Serial("/dev/serial0", 19200, timeout=5)
printer = ThermalPrinter(uart)

@subscription(["$ $ wish text $text would be thermal printed"])
def sub_callback(results):
    claims = []
    claims.append({"type": "retract", "fact": [
        ["variable", ""],
        ["variable", ""],
        ["text", "wish"],
        ["text", "text"],
        ["variable", ""],
        ["text", "would"],
        ["text", "be"],
        ["text", "thermal"],
        ["text", "printed"],
    ]})
    batch(claims)
    for result in results:
        logging.info("PRINTING:")
        logging.info(result["text"].replace(chr(9787), '"'))
        printer.print(result["text"].replace(chr(9787), '"'))
        printer.feed(2)

init(__file__)
