const fs = require('fs');
const { room, myId, run } = require('../helper2')(__filename);

room.on(
  `$ wish $name would be compiled to js`,
  results => {
    console.log("results:")
    console.log(results)
    room.cleanup()
    const name = results[0].name;
    const sourceCode = fs.readFileSync(`src/standalone_processes/${name}`, 'utf8');
    const parsedSourceCode = parse(sourceCode);
    fs.writeFile(
      `src/standalone_processes/${name.replace(".prejs", ".js")}`,
      parsedSourceCode, (err) => {
        if (err) throw err;
        console.error('The file has been saved!');
      }
    );
  }
)

function parse(x) {
  const importPrefix = "const { room, myId, run } = require('../helper2')(__filename);\n\n"
  const runPostfix = "\n\nrun();"

  const whenOtherwiseFunc = s => {
    return s.replace(/when ([^:]*):([\s\S]*)otherwise:\n([\s\S]*$)/g, (match, p1, p2, p3) => {
      const middle = p1.split(",").map(a => a.trim()).join(`",\n        "`)
      return `room.on(\`${middle}\`,\n        results => {\nif (!results) {\n` + p2 + "} else {\n" + p3 + "\n}\n})\n"
    })
  }
  const whenEndFunc = s => {
    return s.replace(/when ([^:]*):([\s\S]*)end\n/g, (match, p1, p2) => {
      const middle = p1.split(",").map(a => a.trim()).join(`",\n        "`)
      return `room.on(\`${middle}\`,\n        results => {\n` + p2 + "\n})\n"
    })
  }
  const whenFunc = s => {
    return s.replace(/when ([^:]*):([\s\S]*$)/g, (match, p1, p2) => {
      const middle = p1.split(",").map(a => a.trim()).join(`",\n        "`)
      return `room.on(\`${middle}\`,\n        results => {` + p2 + "\n})\n"
    })
  }
  const claimFunc = s => {
    return s.replace(/claim ([^\n]*)/g, (match, p1) => {
      return `room.assert(\`${p1}\`)`;
    })
  }
  const retractFunc = s => {
    return s.replace(/retract ([^\n]*)/g, (match, p1) => {
      return `room.retract(\`${p1}\`)`;
    })
  }
  const cleanupFunc = s => {
    return s.replace(/cleanup\n/g, (match, p1) => {
      return `room.cleanup()\n`;
    });
  }

  let s = x;
  s = whenEndFunc(s)
  s = whenOtherwiseFunc(s)
  s = whenFunc(s)
  s = claimFunc(s)
  s = retractFunc(s)
  s = cleanupFunc(s)
  return importPrefix + s + runPostfix;
}

run();
