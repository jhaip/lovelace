const { room, myId, run } = require('../helper2')(__filename);

room.onRaw(`#0280 $ draw graphics $graphics on $`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results) {
            results.forEach(({ graphics }) => {
                let parsedGraphics = JSON.parse(graphics)
                parsedGraphics.forEach(g => {
                    if (g["type"] === "text") {
                        console.log(g["text"])
                    }
                })
            });
        }
        room.subscriptionPostfix();
    })

run();
