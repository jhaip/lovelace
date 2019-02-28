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

const PAPER_WIDTH = 200
const PAPER_HEIGHT = 350
const PAPER_H_MARGIN = 20
const PAPER_V_MARGIN = 20
const ORIGIN_X = 400
const ORIGIN_Y = 900
const x1 = x => ORIGIN_X + x * (PAPER_HEIGHT + PAPER_V_MARGIN);
const x2 = x => ORIGIN_X + x * (PAPER_HEIGHT + PAPER_V_MARGIN) + PAPER_HEIGHT;
const y1 = y => ORIGIN_Y + y * (PAPER_WIDTH + PAPER_H_MARGIN);
const y2 = y => ORIGIN_Y + y * (PAPER_WIDTH + PAPER_H_MARGIN) + PAPER_WIDTH;
const Z = (x, y) => `TL (${x1(x)}, ${y1(y)}) TR (${x2(x)}, ${y1(y)}) BR (${x2(x)}, ${y2(y)}) BL (${x1(x)}, ${y2(y)})`
const sensorScreenLocations = {
  '1': Z(1, 1),
  '2': Z(1, 0),
  '3': 'TL (0, 0) TR (1, 0) BR (1, 1) BL (0, 1)',
  '4': Z(0, 0),
  '5': Z(0, 1)
}

room.on(`$photonId read $value on sensor $sensorId`,
        results => {
  room.subscriptionPrefix(1);
  if (!!results) {
    results.forEach(({ photonId, value, sensorId }) => {
    if (value in rfidValueToPaperId) {
        room.assert(`camera 1 sees paper ${rfidValueToPaperId[value]} at ${sensorScreenLocations[sensorId]} @ 1`);
    }

    });
  }
  room.subscriptionPostfix();
})

room.on(`$photonId read $value on sensor 3`,
  results => {
    room.subscriptionPrefix(1);
    if (!!results) {
      results.forEach(({ photonId, value, sensorId }) => {
        if (value in rfidValueToPaperId) {
          room.assert(`paper 1013 is pointing at paper ${rfidValueToPaperId[value]}`);
        }
      });
    }
    room.subscriptionPostfix();
  })


run();
