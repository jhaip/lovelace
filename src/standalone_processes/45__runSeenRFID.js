const { room, myId, run } = require('../helper2')(__filename);

const rfidValueToPaperId = {
    'TODO': 2000
}

room.on(`$photonId read $value on sensor $sensorId`,
        results => {
  subscriptionPrefix();
  if (!!results) {
    results.forEach(({ photonId, value, sensorId }) => {
    if (value in rfidValueToPaperId) {
        room.assert(`camera 99 sees paper ${paperId} at TL (0, 0) TR (1, 0) BR (1, 1) BL (0, 1) @ 1`)
    }

    });
  }
  subscriptionPostfix();
})


run();
