const { room, myId, run } = require('../helper2')(__filename);

room.on(`$ $ keypad last button pressed is $key`, results => {
  room.cleanup()
  let key = results.length > 0 ? results[0].key : '1';
  if (key == '*' || key == '#') {
    key = '1'
  }
  let targetPaper = 2000 + parseInt(key) - 1
  room.assert(`paper 1013 is pointing at paper ${targetPaper}`)
})

run()
