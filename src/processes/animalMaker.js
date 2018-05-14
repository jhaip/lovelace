// This listens for animals and gives them some attributes, like sight distance

// listens `$ is a $species animal at ($, $)`
// asserts `$species can see ${distance}`

module.exports = async room => {
  if (!room) {
    const Room = require('@living-room/client-js')
    room = new Room()
  }

  room.subscribe(
    [`animalMaker is active`, `$ is a $species animal at ($, $)`],
    async ({ assertions }) => {
      for (const { species: { value: species } } of assertions) {
        if ((await room.select(`${species} can see $`)).length) return

        room.assert(`${species} can see ${Math.random()}`)
      }
    }
  )

  room.assert('animalMaker is active')
}
