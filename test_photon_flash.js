var Particle = require('particle-api-js');
const request = require('request');
const fs = require('fs');
var particle = new Particle();
var token;

var TARGET_DEVICE_ID = '3c002f000e47343432313031';

function downloadFirmwareBinary({ binaryId, auth, stream }) {
    const req = request.get(`/v1/binaries/${binaryId}`);
    req.use(this.agent.prefix);
    req.set({ Authorization: `Bearer ${auth}` });
    req.pipe(stream);
}

particle.login({ username: 'haipjacob@gmail.com', password: 'FILL THIS IN' }).then(
    function (data) {
        token = data.body.access_token;
        console.log("got token");
        console.log(token);

        var devicesPr = particle.listDevices({ auth: token });

        devicesPr.then(
            function (devices) {
                // console.log('Devices: ', devices);
                console.log('Devices: ', devices.body.filter(x => x.connected))

                if (devices.body.filter(x => x.connected && x.id === TARGET_DEVICE_ID).length === 0) {
                    console.log("target device not found", TARGET_DEVICE_ID)
                } else {
                    const url = `https://api.particle.io/v1/devices/${TARGET_DEVICE_ID}?access_token=${token}`;
                    // const formData = {
                    //     file: fs.createReadStream(`/Users/jhaip/Code/lovelace/src/particle-photon/blink/src/blink.ino`),
                    //     // options: {
                    //     //     filename: 'topsecret.jpg',
                    //     // }
                    // }
                    const formData = {
                        file: fs.createReadStream(`/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/dht-sensor.ino`),
                        file1: fs.createReadStream(`/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/HttpClient.h`),
                        file2: fs.createReadStream(`/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/HttpClient.cpp`),
                        file3: fs.createReadStream(`/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/Adafruit_DHT.h`),
                        file4: fs.createReadStream(`/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/Adafruit_DHT.cpp`)
                    }
                    var req = request.put({url, formData}, function (err, resp, body) {
                        if (err) {
                            console.log('Error!');
                        } else {
                            console.log("successful compile");
                            console.log(body);
                        }
                    });
                }
            },
            function (err) {
                console.log('List devices call failed: ', err);
            }
        );
    },
    function (err) {
        console.log('Could not log in.', err);
    }
);