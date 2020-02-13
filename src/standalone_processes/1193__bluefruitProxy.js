
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

const BLUEFRUIT_BUTTON_SERVICE = 'adaf0600c33242a893bd25e905756cb8';
const BLUEFRUIT_LIGHT_SENSOR_SERVICE = 'adaf0300c33242a893bd25e905756cb8';
const SERVICE_CHARACTERISTICS = {
    [BLUEFRUIT_BUTTON_SERVICE]: 'adaf0601c33242a893bd25e905756cb8',
    [BLUEFRUIT_LIGHT_SENSOR_SERVICE]: 'adaf0301c33242a893bd25e905756cb8'
}
const SERVICE_NAMES = {
    [BLUEFRUIT_BUTTON_SERVICE]: 'buttons',
    [BLUEFRUIT_LIGHT_SENSOR_SERVICE]: 'light_sensor'
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
        const supportedServices = [BLUEFRUIT_BUTTON_SERVICE, BLUEFRUIT_LIGHT_SENSOR_SERVICE];
        peripheral.discoverServices(supportedServices, function (error, services) {
            services.forEach(service => {
                var serviceUuid = `${service.uuid}`;
                console.log(`discovered ${SERVICE_NAMES[serviceUuid]} service ${service}`);

                service.discoverCharacteristics([SERVICE_CHARACTERISTICS[serviceUuid]], function (error, characteristics) {
                    var characteristic = characteristics[0];
                    console.log(`discovered ${SERVICE_NAMES[serviceUuid]} characteristic`);

                    characteristic.on('data', function (data, isNotification) {
                        console.log(`${SERVICE_NAMES[serviceUuid]} is now: ${CONVERSION_FUNC[serviceUuid](data)}`);
                    });

                    // to enable notify
                    characteristic.subscribe(function (error) {
                        console.log(`${SERVICE_NAMES[serviceUuid]} notification on`);
                    });
                });
            });
        });
    });
});

// Run with:
// which node
// sudo the-node-from-above 1193__bluefruitProxy.js

// run();
