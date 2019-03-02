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
  'e21eef27': 6,
  '7bdbe359': 1013
}

const PAPER_WIDTH = 260
const PAPER_HEIGHT = 375
const PAPER_H_MARGIN = 150
const PAPER_V_MARGIN = 120
const ORIGIN_X = 360;
const ORIGIN_Y = 820;
const x1 = x => x * (PAPER_WIDTH + PAPER_H_MARGIN);
const x2 = x => x * (PAPER_WIDTH + PAPER_H_MARGIN) + PAPER_WIDTH;
const y1 = y => y * (PAPER_HEIGHT + PAPER_V_MARGIN);
const y2 = y => y * (PAPER_HEIGHT + PAPER_V_MARGIN) + PAPER_HEIGHT;
const Z = (x, y) => {
  return `TL (${ORIGIN_X + y1(y)}, ${ORIGIN_Y - x1(x)}) ` + 
    `TR (${ORIGIN_X + y1(y)}, ${ORIGIN_Y - x2(x)}) ` + 
    `BR (${ORIGIN_X + y2(y)}, ${ORIGIN_Y - x2(x)}) ` + 
    `BL (${ORIGIN_X + y2(y)}, ${ORIGIN_Y - x1(x)})`
}
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
