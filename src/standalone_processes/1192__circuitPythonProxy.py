from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen, get_my_id_str
import serial
import logging
import time

write_buffer = []

@subscription(["$ $ wish circuit playground neopixel $i had color $r $g $b"])
def melody_callback(results):
    global write_buffer
    if results:
        for result in results:
            write_buffer.append("LIGHT,{},{},{},{}\n".format(
                result["i"], result["r"], result["g"], result["b"]).encode("utf-8"))


init(__file__, skipListening=True)


with serial.Serial('/dev/ttyUSB0', 115200, timeout=1.0) as ser:
    ser.reset_input_buffer()
    ser.reset_output_buffer()
    while True:
        received_msg = True
        while received_msg:
            logging.info("checking for messages from room")
            received_msg = listen(blocking=False)
        # Send new messages if there are any
        if len(write_buffer) > 0:
            logging.info("writing to serial:")
            logging.info(write_buffer)
            for line in write_buffer:
                ser.write(line)
            write_buffer = []
        # Receive serial messages
        claims = [
            {"type": "retract", "fact": [["id", get_my_id_str()], ["id", "0"], ["postfix", ""]]}
        ]
        logging.info("reading serial lines")
        lines = ser.readlines()  # used the serial timeout specified above
        logging.info("done reading serial lines")
        for line in lines:
            # Example: line = b'BUTTON_A:1\n'
            try:
                parsed_line = line.rstrip().split(b":")
                if len(parsed_line) is 2:
                    prefix = parsed_line[0].decode("utf-8")
                    value = parsed_line[1]
                    if prefix == 'BUTTON_A' or prefix == 'BUTTON_B' or prefix == 'LIGHT':
                        claims.append({"type": "claim", "fact": [
                            ["id", get_my_id_str()],
                            ["id", "0"],
                            ["text", "circuit"],
                            ["text", "playground"],
                            ["text", prefix],
                            ["text", "has"],
                            ["text", "value"],
                            ["integer", value.decode("utf-8")],
                        ]})
                    else:
                        logging.info("Ignoring message: {}".format(line))
            except:
                logging.error("Unexpected error:", sys.exc_info()[0])
        if claims:
            batch(claims)
        time.sleep(0.1)
