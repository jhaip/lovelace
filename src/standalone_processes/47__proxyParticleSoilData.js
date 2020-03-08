const { room, myId, run } = require('../helper2')(__filename);
const path = require('path');
const Particle = require('particle-api-js');
const request = require('request');
const fs = require('fs');
const particle = new Particle();

const LOGIN_INFO = { username: 'haipjacob@gmail.com', password: process.env.PARTICLE_PASSWORD };

var token;

function handleRequest() {
    if (!token) {
        console.log("logging in...")
        particle
            .login(LOGIN_INFO)
            .then(data => {
                token = data.body.access_token;
                console.log("got token", token);
                handleRequest();
            })
    } else {
        console.log("getting event stream!")
        particle.getEventStream({ name: 'rgb_lights', auth: token }).then(function (stream) {
            stream.on('event', function (data) {
                console.log("Event: ", data);
                // data looks like this:
                // {
                // "name":"Uptime",
                // "data":"5:28:54",
                // "ttl":"60",
                // "published_at":"2014-MM-DDTHH:mm:ss.000Z",
                // "coreid":"012345678901234567890123"
                // }
                if (data.name === "TestSensor") {
                    const dataValueArr = data.data.split(":");
                    if (dataValueArr.length === 2) {
                        const sensorName = dataValueArr[0];
                        const sensorValue = dataValueArr[1].toInt();
                        room.assert(`Photon says "${sensorName}" is ${sensorValue}`)
                    }
                }
            });
            stream.on('error', function (error) {
                console.log('Error in handler', error);
            });
        })
        .catch(err => {
            console.error("ERROR:")
            console.error(err);
            // TODO: make a claim about the error
        });
    }
}

handleRequest();

run();