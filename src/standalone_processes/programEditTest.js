const Room = require('@living-room/client-js')
const room = new Room()

room.retract(`wish testy.js has source code $sourceCode`)

let sourceCode = `
const Room = require('@living-room/client-js')
const room = new Room()

// comment

room.assert('hello from way inside programEditTest')
`;

sourceCode = sourceCode.replace(/\n/g, "\\n")

room.assert(`wish testy.js has source code \"${sourceCode}\"`)
