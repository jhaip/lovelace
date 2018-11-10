const { room, myId, scriptName, MY_ID_STR, run } = require('../helper2')(__filename);

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
    `#${MY_ID_STR} ping $time`,
    results => {
        const pingTime = new Date(parseInt(results[0].time))
        const latencyMs = (new Date()) - pingTime
        console.log("LATENCY (ms):", latencyMs)
        room.cleanup()
        room.assert(`measured latency ${latencyMs} ms at`, ["text", (new Date()).toUTCString()])
        setTimeout(sendPing, delayBetweenMeasurementsMs)
    }
)

run()
sendPing()