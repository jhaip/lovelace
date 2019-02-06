const fs = require('fs');
const { room, myId, run } = require('../helper2')(__filename);

room.onGetSource('wisherId',
  `wish $name would be compiled to js`,
  results => {
    console.log("results:")
    console.log(results)
    room.cleanup()
    results.forEach(({ wisherId, name }) => {
      const sourceCode = fs.readFileSync(`src/standalone_processes/${name}`, 'utf8');
      const parsedSourceCode = parse(sourceCode);
      fs.writeFile(
        `src/standalone_processes/${name.replace(".prejs", ".js")}`,
        parsedSourceCode, (err) => {
          if (err) throw err;
          console.error('The file has been saved!');
          room.retractFromSource(wisherId, `wish`, ["text", name], `would be compiled to js`)
          room.flush();
        }
      );
    })
  }
)

function parse(x) {
  const importPrefix = "const { room, myId, run } = require('../helper2')(__filename);\n\n"
  const runPostfix = "\n\nrun();"

  function onlyUnique(value, index, self) {
    return self.indexOf(value) === index;
  }

  const getUniqueVariables = s => {
    const variables = s.match(/\$([a-zA-Z0-9]+)/g);
    if (variables) {
      return variables.map(x => x.slice(1)).filter(onlyUnique)
    }
    return [];
  }

  const whenOtherwiseEndFunc = s => {
    return s.replace(/when ([^:]*):([\s\S]+?)\notherwise:\n([\s\S]+?\n)end(\n|$)/g, (match, p1, p2, p3) => {
      const middle = p1.split(",\n").map(a => a.trim()).join(`\`,\n        \``)
      const variables = getUniqueVariables(p1);
      return `room.on(\`${middle}\`,\n        results => {\nsubscriptionPrefix();\nif (!!results) {\n  results.forEach(({ ${variables.join(", ")} }) => {\n` + p2 + "\n  });\n} else {\n" + p3 + "}\nsubscriptionPostfix();\n})\n"
    })
  }
  const whenOtherwiseFunc = s => {
    return s.replace(/when ([^:]*):([\s\S]+?)\notherwise:\n([\s\S]+?$)/g, (match, p1, p2, p3) => {
      const middle = p1.split(",\n").map(a => a.trim()).join(`\`,\n        \``)
      const variables = getUniqueVariables(p1);
      return `room.on(\`${middle}\`,\n        results => {\nsubscriptionPrefix();\nif (!!results) {\n  results.forEach(({ ${variables.join(", ")} }) => {\n` + p2 + "\n  });\n} else {\n" + p3 + "\n}\nsubscriptionPostfix();\n})\n"
    })
  }
  const whenEndFunc = s => {
    return s.replace(/when ([^:]*):([\s\S]+?\n)end(\n|$)/g, (match, p1, p2) => {
      const middle = p1.split(",\n").map(a => a.trim()).join(`\`,\n        \``)
      return `room.on(\`${middle}\`,\n        results => {\nsubscriptionPrefix();\n` + p2 + "\nsubscriptionPostfix();\n})\n"
    })
  }
  const whenFunc = s => {
    return s.replace(/when ([^:]*):([\s\S]+?$)/g, (match, p1, p2) => {
      const middle = p1.split(",\n").map(a => a.trim()).join(`\`,\n        \``)
      return `room.on(\`${middle}\`,\n        results => {\nsubscriptionPrefix();\n` + p2 + "\nsubscriptionPostfix();\n})\n"
    })
  }
  const claimFunc = s => {
    return s.replace(/claim ([^\n]*)/g, (match, p1) => {
      return `room.assert(\`${p1}\`)`;
    })
  }
  const retractFunc = s => {
    return s.replace(/retract ([^\n]*)/g, (match, p1) => {
      return `room.retractAll(\`${p1}\`)`;
    })
  }
  const cleanupFunc = s => {
    return s.replace(/cleanup\n/g, (match, p1) => {
      return `room.cleanup()\n`;
    });
  }

  let s = x;
  s = whenOtherwiseEndFunc(s)
  s = whenEndFunc(s)
  s = whenOtherwiseFunc(s)
  s = whenFunc(s)
  s = claimFunc(s)
  s = retractFunc(s)
  s = cleanupFunc(s)
  return importPrefix + s + runPostfix;
}

run();
