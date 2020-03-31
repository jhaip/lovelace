const { room, myId, run } = require('../helper2')(__filename);

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
  'Photon2c001b000347343233323032': {
    '1': Z(0, 0),
    '2': 'TL (0, 0) TR (1, 0) BR (1, 1) BL (0, 1)',
    '3': Z(1, 0),
    '4': Z(1, 1),
    '5': Z(0, 1)
  },
  'Photon200038000747343232363230': {
    '3': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)'
  },
  'Photon400035001547343433313338': {
    '1': 'TL (0, 0) TR (1, 0) BR (1, 1) BL (0, 1)'
  },
  'Photon400035001547343433313338': {
    '1': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '3': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)'
  },
  'ArgonBLE': {
    '1': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '2': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '3': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '4': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '5': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
  },
  'ArduinoUSB': {
    '0': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '1': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '2': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
    '3': 'TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080)',
  },
}

room.on(`$photonId read $value on sensor $sensorId`,
        `paper $paperId has RFID $value`,
        results => {
  room.subscriptionPrefix(1);
  if (!!results) {
    results.forEach(({ photonId, value, sensorId, paperId }) => {
      room.assert(`camera 1 sees paper ${paperId} at ${sensorScreenLocations[photonId][sensorId]} @ 1`);

    });
  }
  room.subscriptionPostfix();
})

room.on(`Photon200038000747343232363230 read $value on sensor 1`,
        `paper $paperId has RFID $value`,
  results => {
    room.subscriptionPrefix(2);
    if (!!results) {
      results.forEach(({ value, paperId }) => {
        room.assert(`paper 1013 is pointing at paper ${paperId}`);
      });
    }
    room.subscriptionPostfix();
  })

room.on(`ArgonBLE read $value on sensor 4`,
  `paper $paperId has RFID $value`,
  results => {
    room.subscriptionPrefix(3);
    if (!!results) {
      results.forEach(({ value, paperId }) => {
        room.assert(`paper 1013 is pointing at paper ${paperId}`);
      });
    }
    room.subscriptionPostfix();
  })

// room.on(`Photon200038000747343232363230 read $value on sensor 3`,
//         `paper $paperId has RFID $value`,
//   results => {
//     room.subscriptionPrefix(3);
//     if (!!results) {
//       results.forEach(({ value, paperId }) => {
//         room.assert(`wish display 1700 only showed ${paperId}`);
//       });
//     }
//     room.subscriptionPostfix();
//   })


run();
