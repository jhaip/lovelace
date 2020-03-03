const { room, myId, run } = require('../helper2')(__filename);
const request = require('request');

function fetchWeather() {
    request(
        `https://api.darksky.net/forecast/ff4210a6ee0e933946c817939138eb1f/42.3601,-71.0589?exclude=minutely,hourly,alerts,flags`,
        { json: true },
        (err, res, body) => {
            room.cleanup();
            room.assert(`weather forecast updated at ${(new Date()).getTime()}`)
            if (err) {
                room.assert(`weather forecast had error "${err}"`)
            } else if (!res || res.statusCode !== 200) {
                room.assert(`weather forecast had error "${res && res.statusCode}"`)
            } else {
                room.assert(`current weather is ${Math.floor(body.currently.temperature)} F and ${body.currently.icon}`)
                body.daily.data.forEach(v => {
                    const dateIsoString = (new Date(v.time * 1000)).toISOString()
                    room.assert(
                        `weather forecast for "${dateIsoString}" is ` +
                        `low ${Math.floor(v.temperatureLow)} F high ${Math.floor(v.temperatureHigh)} F and ` +
                        `${v.icon} with "${Math.floor(100 * v.precipProbability)}%" chance of ${v.precipType}`
                    )
                })
            }
            room.flush();
            setTimeout(fetchWeather, 1000 * 60 * 5)
        }
    );
}

fetchWeather();

run();
