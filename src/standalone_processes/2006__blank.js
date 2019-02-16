const { room, myId, run } = require('../helper2')(__filename);

room.on(`$ keyboard $ typed special key $key @ $time`, results => {
  if (results.length === 0) return
  room.cleanup()
  let key = results[0].key
  room.assert(`wish I was labeled LAST KEY PRESSED: special key ${key}`)
})

room.on(`$ keyboard $ typed key $key @ $time`, results => {
  if (results.length === 0) return
  room.cleanup()
  let key = results[0].key
  room.assert(`wish I was labeled LAST KEY PRESSED: key ${key}`)
})

run()