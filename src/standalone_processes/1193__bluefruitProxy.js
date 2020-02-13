
// const { room, myId, MY_ID_STR, run } = require('../helper2')(__filename);

// var noble = require('noble');   //noble library
var noble = require('@abandonware/noble');

// uuids are easier to read with dashes
// this helper removes dashes so comparisons work
var uuid = function (uuid_with_dashes) {
    return uuid_with_dashes.replace(/-/g, '');
};
var strippedMAC = function (mac) {
    return mac.replace(/-/g, '').replace(/:/g, '');
}

// https://learn.adafruit.com/bluefruit-playground-app/ble-services
const BLUEFRUIT_BUTTON_SERVICE = 'adaf0600c33242a893bd25e905756cb8';
const BLUEFRUIT_LIGHT_SENSOR_SERVICE = 'adaf0300c33242a893bd25e905756cb8';
const BLUEFRUIT_TONE_SERVICE = 'adaf0c00c33242a893bd25e905756cb8';
const BLUEFRUIT_NEOPIXEL_SERVICE = 'adaf0900c33242a893bd25e905756cb8';
const SERVICE_CHARACTERISTICS = {
    [BLUEFRUIT_BUTTON_SERVICE]: ['adaf0601c33242a893bd25e905756cb8'],
    [BLUEFRUIT_LIGHT_SENSOR_SERVICE]: ['adaf0301c33242a893bd25e905756cb8'],
    [BLUEFRUIT_TONE_SERVICE]: ['adaf0c01c33242a893bd25e905756cb8'],
    [BLUEFRUIT_NEOPIXEL_SERVICE]: [
        'adaf0901c33242a893bd25e905756cb8',
        'adaf0902c33242a893bd25e905756cb8',
        'adaf0903c33242a893bd25e905756cb8',
        'adaf0904c33242a893bd25e905756cb8'
    ]
}
const SERVICE_NAMES = {
    [BLUEFRUIT_BUTTON_SERVICE]: 'buttons',
    [BLUEFRUIT_LIGHT_SENSOR_SERVICE]: 'light_sensor',
    [BLUEFRUIT_TONE_SERVICE]: 'tone',
    [BLUEFRUIT_NEOPIXEL_SERVICE]: 'neopixel'
}
const CONVERSION_FUNC = {
    [BLUEFRUIT_BUTTON_SERVICE]: data => data.readUInt8(0),
    [BLUEFRUIT_LIGHT_SENSOR_SERVICE]: data => data.readFloatLE(0)
}

// when the radio turns on, start scanning:
noble.on('stateChange', function scan(state) {
    if (state === 'poweredOn') {    // if the radio's on, scan for this service
        noble.startScanning([], false);
    }
});

function playTone(characteristic) {
    // characteristic.write(Buffer.from([0xB8, 0x01, 0xFF, 0x00, 0x00, 0x00]), false, function (error) {
    //     console.log('playing tone');
    //     console.log(error);
    // });
    // Uint16 frequency, reverse byte order. Ex: 0xB1 0x01 = 440
    // Uint32 milliseconds to play, reverse byte order. Ex: 0xFF 0x00 0x00 0x00 = 255
    characteristic.write(Buffer.from([0x00, 0x00, 0x00, 0x00, 0x00, 0x00]), true, function (error) {
        console.log('cleared tone');
        console.log(error);
    });
    setTimeout(function() {
        // 0xB1 0x01 = Uint16 440 Hz tone
        characteristic.write(Buffer.from([0xB8, 0x01, 0x00, 0x00, 0x00, 0x00]), true, function (error) {
            console.log('playing tone');
            console.log(error);
        });
        setTimeout(function () {
            characteristic.write(Buffer.from([0x00, 0x00, 0x00, 0x00, 0x00, 0x00]), true, function (error) {
                console.log('cleared tone');
                console.log(error);
            });
        }, 200);
    }, 100);
}

// if you discover a peripheral with the appropriate service, connect:
// noble.on('discover', self.connect);
noble.on('discover', function (peripheral) {
    console.log(`inside discover ${peripheral.address}`)
    if (strippedMAC(peripheral.address) !== strippedMAC('d1:d3:b6:0c:9b:95')) {
        return;
    } else {
        console.log("FOUND BLUEFRUIT!");
    }
    peripheral.connect(function (error) {
        console.log('connected to peripheral: ' + peripheral.uuid);
        // adaf0600c33242a893bd25e905756cb8 are the bluefruit buttons
        const supportedServices = [BLUEFRUIT_BUTTON_SERVICE, BLUEFRUIT_LIGHT_SENSOR_SERVICE, BLUEFRUIT_TONE_SERVICE, BLUEFRUIT_NEOPIXEL_SERVICE];
        peripheral.discoverServices(supportedServices, function (error, services) {
            services.forEach(service => {
                var serviceUuid = `${service.uuid}`;
                console.log(`discovered ${SERVICE_NAMES[serviceUuid]} service ${service}`);

                service.discoverCharacteristics(SERVICE_CHARACTERISTICS[serviceUuid], function (error, characteristics) {
                    console.log(`discovered ${SERVICE_NAMES[serviceUuid]} characteristics`);
                    const serviceName = SERVICE_NAMES[serviceUuid];
                    if (serviceName === 'tone') {
                        var characteristic = characteristics[0];
                        playTone(characteristic);
                    } else if (serviceName === 'neopixel') {
                        // var characteristic = characteristics[0];
                        characteristics.forEach(characteristic => {
                            // console.log(characteristic);
                            var characteristicUuid = `${characteristic.uuid}`;
                            if (characteristicUuid === SERVICE_CHARACTERISTICS[BLUEFRUIT_NEOPIXEL_SERVICE][0]) {
                                characteristic.write(Buffer.from([0x08]), true, function (error) {
                                    console.log('wrote neopixel 0');
                                    console.log(error);
                                });
                            }
                            if (characteristicUuid === SERVICE_CHARACTERISTICS[BLUEFRUIT_NEOPIXEL_SERVICE][1]) {
                                characteristic.write(Buffer.from([0x00]), true, function (error) {
                                    console.log('wrote neopixel 1');
                                    console.log(error);
                                });
                            }
                            if (characteristicUuid === SERVICE_CHARACTERISTICS[BLUEFRUIT_NEOPIXEL_SERVICE][2]) {
                                const msg = [
                                    0x00, 0x00,
                                    0x01,
                                    0x00, 0x00, 0x00, // Neopixel 0 color in GRB
                                    0x05, 0x00, 0x00, // Neopixel 1 color in GRB
                                    0x10, 0x00, 0x00, // ...
                                    0x00, 0x05, 0x00,
                                    0x00, 0x10, 0x00,
                                    0x00, 0x00, 0x05,
                                    0x00, 0x00, 0x10,
                                    0x00, 0x00, 0x00,
                                    0x00, 0x00, 0x00,
                                    0x00, 0x00, 0x00
                                ]
                                characteristic.write(Buffer.from(msg), true, function (error) {
                                    console.log('wrote neopixel 2');
                                    console.log(error);
                                });
                            }
                            if (characteristicUuid === SERVICE_CHARACTERISTICS[BLUEFRUIT_NEOPIXEL_SERVICE][3]) {
                                // set buffer size to 30
                                characteristic.write(Buffer.from([0x1E, 0x00]), true, function (error) {
                                    console.log('wrote neopixel 3');
                                    console.log(error);
                                });
                            }
                        })
                    } else {
                        var characteristic = characteristics[0];

                        characteristic.on('data', function (data, isNotification) {
                            console.log(`${SERVICE_NAMES[serviceUuid]} is now: ${CONVERSION_FUNC[serviceUuid](data)}`);
                        });

                        // to enable notify
                        characteristic.subscribe(function (error) {
                            console.log(`${SERVICE_NAMES[serviceUuid]} notification on`);
                        });
                    }
                });
            });
        });
    });
});

// Run with:
// which node
// sudo the-node-from-above 1193__bluefruitProxy.js

// run();
