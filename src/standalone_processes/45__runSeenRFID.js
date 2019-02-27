const { room, myId, run } = require('../helper2')(__filename);

const rfidValueToPaperId = {
  'f26a0c2e': 2000,
  'f238222e': 2001,
  '80616ea3': 2002,
  'd07911a3': 2003,
  '91b4d108': 2004,
  '53825027': 2005,
  '10af78a3': 2006,
  '7341a727': 2007,
  '2574c72d': 2008,
  'b680cc21': 2009,
  'd01ff625': 5,
  'e21eef27': 6
}

room.on(`$photonId read $value on sensor $sensorId`,
        results => {
  room.subscriptionPrefix(1);
  if (!!results) {
    results.forEach(({ photonId, value, sensorId }) => {
    if (value in rfidValueToPaperId) {
        room.assert(`camera 99 sees paper ${rfidValueToPaperId[value]} at TL (0, 0) TR (1, 0) BR (1, 1) BL (0, 1) @ 1`);
        if (String(sensorId) === String(3)) {
            room.assert(`paper 1013 is pointing at paper ${rfidValueToPaperId[value]}`);
        }
    }

    });
  }
  room.subscriptionPostfix();
})


run();
