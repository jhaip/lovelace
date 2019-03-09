const { room, myId, scriptName, MY_ID_STR, run } = require('../helper2')(__filename);
const fs = require('fs');

const stream = fs.createWriteStream("files/latency-ms-log.txt", { flags: 'a' });

var lastSentPing
const serverTimeoutMs = 5000
const delayBetweenMeasurementsMs = 2000

function sendPing() {
    let currentTimeMs = (new Date()).getTime()
    lastSentPing = currentTimeMs
    room.assert(`ping ${currentTimeMs}`)
    room.flush()
    setTimeout(() => {
        if (lastSentPing <= currentTimeMs) {
            console.error(`SERVER TIMEOUT - NO PING RESPONSE IN ${serverTimeoutMs} ms! ${(new Date()).toUTCString()}`)
        }
    }, serverTimeoutMs)
}

room.on(
    `ping $time`,
    results => {
        const pingTime = new Date(parseInt(results[0].time))
        const latencyMs = (new Date()) - pingTime
        console.log("LATENCY (ms):", latencyMs)
        room.cleanup()
        room.assert(`measured latency ${latencyMs} ms at`, ["text", (new Date()).toUTCString()])
        stream.write(`${(new Date()).toUTCString()},${latencyMs}\n`);
        setTimeout(sendPing, delayBetweenMeasurementsMs)
    }
)

run()
sendPing()