const { room, myId, run } = require('../helper2')(__filename);
const Particle = require('particle-api-js');
const request = require('request');
const fs = require('fs');
const particle = new Particle();

const LOGIN_INFO = { username: 'haipjacob@gmail.com', password: 'TODO FILL IN' };

var token;

room.on(`wish $code runs on photon $photonId`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results) {
            results.forEach(({ code, photonId }) => {
                if (code.includes(`#include "HttpClient.h"`)) {
                    console.log("adding in HTTP definitions to code");
                    const firstBlankLineIndex = code.indexOf("\n\n") + 1;
                    const http_setup_code = `
                    HttpClient http;

// Headers currently need to be set at init, useful for API keys etc.
http_header_t headers[] = {
    {"Content-Type", "application/json"},
    {"Accept", "application/json"},
    {"Accept", "*/*"},
    {NULL, NULL} // NOTE: Always terminate headers will NULL
};

http_request_t request;
http_response_t response;

String myID = System.deviceID();

void publishValueMessage(char body[])
{
    request.ip = {10, 0, 0, 185};
    request.port = 5000;
    request.path = "/cleanup-claim";
    request.body = str;
    Serial.println(request.body);
    http.post(request, response, headers);
    Serial.print("Application>\tResponse status: ");
    Serial.println(response.status);
}
                    `;
                    code = [code.slice(0, firstBlankLineIndex), http_setup_code, code.slice(firstBlankLineIndex)].join('');
                }

                particle
                    .login(LOGIN_INFO)
                    .then(data => {
                        token = data.body.access_token;
                        console.log("got token", token);
                        return particle.listDevices({ auth: token });
                    })
                    .then(devices => {
                        console.log('Devices: ', devices.body.filter(x => x.connected))
                        if (devices.body.filter(x => x.connected && x.id === photonId).length === 0) {
                            return new Promise((resolve, reject) => {
                                throw new Error(`target device not found ${photonId}`);
                                resolve(true);
                            });
                        }
                        console.log("sending code:")
                        console.log(code);
                        const url = `https://api.particle.io/v1/devices/${photonId}?access_token=${token}`;
                        const formData = {
                            file: {
                                value: Buffer.from(code),
                                options: {
                                    filename: 'code.ino',
                                    contentType: 'text/plain'
                                }
                            }
                        }
                        let currentFileIndex = 1;
                        if (code.includes(`#include "HttpClient.h"`)) {
                            formData[`file${currentFileIndex}`] = `/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/HttpClient.h`;
                            currentFileIndex += 1;
                            formData[`file${currentFileIndex}`] = `/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/HttpClient.cpp`;
                            currentFileIndex += 1;
                        }
                        if (code.includes(`#include "Adafruit_DHT.h"`)) {
                            formData[`file${currentFileIndex}`] = `/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/Adafruit_DHT.h`;
                            currentFileIndex += 1;
                            formData[`file${currentFileIndex}`] = `/Users/jhaip/Code/lovelace/src/particle-photon/dht-sensor/Adafruit_DHT.cpp`;
                            currentFileIndex += 1;
                        }
                        console.log("using form data:")
                        console.log(formData);
                        var req = request.put({ url, formData }, function (err, resp, body) {
                            if (err) {
                                return new Promise((resolve, reject) => {
                                    throw new Error(err);
                                    resolve(true);
                                });
                            }
                            console.log("successful compile");
                            console.log(body);
                            return new Promise((resolve, reject) => resolve(true));
                        });
                    })
                    .catch(err => {
                        console.error("ERROR:")
                        console.error(err);
                        // TODO: make a claim about the error
                    });
            });
        }
        room.subscriptionPostfix();
    })