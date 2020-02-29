
const { room, myId, MY_ID_STR, run } = require('../helper2')(__filename);

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
const ARGON_RFID_SERVICE = uuid('4677062c-ad02-4034-9abf-98581772427c');
const ARGON_RFID_CHARACTERISTIC = uuid('dc13b36a-3499-46b0-ac11-5ac0173c4cc5');

var active_tone_characteristic = null;
var active_neopixel_characteristic = null;

// when the radio turns on, start scanning:
noble.on('stateChange', function scan(state) {
    if (state === 'poweredOn') {    // if the radio's on, scan for this service
        noble.startScanning([], false);
    }
});

function connect(peripheral) {
    const supportedServices = [ARGON_RFID_SERVICE];
    peripheral.discoverServices(supportedServices);
    console.log("started discovering services");
    peripheral.once('servicesDiscover', function (services) {
        services.forEach(service => {
            var serviceUuid = `${service.uuid}`;
            console.log(`discovered service ${service}`);

            service.discoverCharacteristics([ARGON_RFID_CHARACTERISTIC], function (error, characteristics) {
                console.log(`discovered characteristics`);
                var characteristic = characteristics[0];

                characteristic.on('data', function (data, isNotification) {
                    const characteristicValue = data.toString('utf8')
                    console.log(`value is now: ${characteristicValue}`);
                    // room.retractMine(`circuit playground "LIGHT" has value $`);
                    // room.assert(`circuit playground "LIGHT" has value ${lightValue}`);
                    // room.flush();
                });

                // to enable notify
                characteristic.subscribe(function (error) {
                    console.log(`notification on`);
                });
            });
        });
    });
}

// if you discover a peripheral with the appropriate service, connect:
// noble.on('discover', self.connect);
noble.on('discover', function (peripheral) {
    console.log(`inside discover ${peripheral.address}`)
    // TODO: update MAC
    if (strippedMAC(peripheral.address) !== strippedMAC('d1:d3:b6:0c:9b:95')) {
        return;
    } else {
        console.log("FOUND ARGON!");
        console.log(peripheral);
    }

    peripheral.connect();
    peripheral.on('connect', function () {
        noble.stopScanning();
        console.log('connected to peripheral: ' + peripheral.uuid);
        connect(peripheral);
    });
    peripheral.on('disconnect', function () {
        console.log("PERIPHERAL DISCONNECTED")
        setTimeout(() => {
            console.log("attempting to reconnect");
            peripheral.connect();
        }, 1000);
    });
});

// Run with:
// which node
// sudo the-node-from-above 1193__bluefruitProxy.js

room.cleanup();
room.retractRaw(`subscription #1851 %`);

run();
