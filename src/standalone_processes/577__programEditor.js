const fs = require('fs');
const { room, myId, scriptName, run } = require('../helper2')(__filename);

room.on(
  `$wisherId wish $name has source code $sourceCode`,
  results => {
    console.error(results);
    results.forEach(({ wisherId, name, sourceCode }) => {
      console.error("ON WISH SOURCE CODE")
      console.error(wisherId)
      console.error(name)
      console.error(sourceCode)
      if (!name.includes('.py') && !name.includes('.js')) {
        name += '.js'
      }
      sourceCode = sourceCode.replace(new RegExp(String.fromCharCode(9787), 'g'), String.fromCharCode(34))
      console.log('debug:::')
      console.log(`#${wisherId} wish "${name}" has source code $`)
      room.retract(`#${wisherId} wish`, ["text", name], `has source code $`)
      room.retract(`$`, ["text", name], `has source code $`);
      fs.writeFile(`src/standalone_processes/${name}`, sourceCode, (err) => {
        if (err) throw err;
        console.error('The file has been saved!');
        room.assert(["text", name], `has source code`, ["text", sourceCode]);
      });
    })
  }
);
