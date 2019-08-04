const { room, myId, run } = require('../helper2')(__filename);

room.on(`laser seen at $x $y @ $t`,
    `camera $ sees paper $paper at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $t2`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results) {
            results.forEach(({ x, y, t, paper, x1, y1, x2, y2, x3, y3, x4, y4, t2 }) => {
                // Code from https://github.com/substack/point-in-polygon/blob/master/index.js
                let inside = false;
                let vs = [[x1, y1], [x2, y2], [x3, y3], [x4, y4]];
                for (var i = 0, j = vs.length - 1; i < vs.length; j = i++) {
                    let xi = vs[i][0], yi = vs[i][1];
                    let xj = vs[j][0], yj = vs[j][1];
                    let intersect = ((yi > y) != (yj > y))
                        && (x < (xj - xi) * (y - yi) / (yj - yi) + xi);
                    if (intersect) inside = !inside;
                }
                if (inside) {
                    room.assert(`laser in paper ${ paper }`)
                }

            });
        }
        room.subscriptionPostfix();
    })


run();
