const { room, myId, run } = require('../helper2')(__filename);
const sensorOrder = [4, 5, 3, 1];

var lastStack = [];
var turtles = [];

room.onRaw(`$ $ ArgonBLE read $value on sensor $sensorId`,
    `$ $ paper $paperNumber has RFID $value`,
    `$ $ paper $paperNumber has id $programId`,
    `$programId $ I am a $cardType card`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results) {
            let cards = ["", "", "", ""];
            results.forEach(({ value, sensorId, paperNumber, programId, cardType }) => {
                const cardPosition = sensorOrder.indexOf(sensorId);
                if (cardPosition >= 0) {
                    cards[cardPosition] = cardType;
                }
            });
            let stack = cards.filter(card => card !== "");
            console.log("NEW STACK:")
            console.log(stack);
            if (JSON.stringify(stack) !== JSON.stringify(lastStack)) {
                lastStack = stack;
                turtles = [];
                if (stack.length > 0 && stack[0] === "turtle") {
                    const hasSpiral = stack.indexOf("spiral") > 0;
                    const hasTail = stack.indexOf("pen") > 0;
                    turtles.push({
                        x: 0,
                        y: 0,
                        heading: Math.random() * 2.0 * Math.PI,
                        speed: 3,
                        movementType: hasSpiral ? "spiral" : "random",
                        hasTail: hasTail,
                        trail: []
                    });
                }
            }
        }
        room.subscriptionPostfix();
    })

setInterval(() => {
    console.log(turtles);
    // update
    for (let i = 0; i < turtles.length; i += 1) {
        if (turtles[i].movementType === "random") {
            turtles[i].heading = Math.random() * 2.0 * Math.PI;
        } else if (turtles[i].movementType === "spiral") {
            turtles[i].heading += 1.0 / (Math.PI * 2.0);
        }
        if (turtles[i].hasTail) {
            turtles[i].tail.push([turtles[i].x, turtles[i].y, 0, 0, 255]);
            turtles[i].tail = turtles[i].tail.slice(0, 1000); // limits length of tail
        }
        turtles[i].x += turtles[i].speed * Math.cos(turtles[i].heading);
        turtles[i].y += turtles[i].speed * Math.sin(turtles[i].heading);
    }
    // draw
    room.cleanup();
    let ill = room.newIllumination()
    ill.push();
    ill.translate(300, 300);
    ill.fill("green")
    for (let i = 0; i < turtles.length; i += 1) {
        for (let t = 0; t < turtles[i].tail.length; t += 1) {
            const tailPoint = turtles[i].tail[t];
            ill.push();
            ill.fill(tailPoint[2], tailPoint[3], tailPoint[4])
            ill.ellipse(-10 + tailPoint[0], -10 + tailPoint[1], 20, 20);
            ill.pop();
        }
        ill.push();
        ill.translate(turtles[i].x, turtles[i].y);
        ill.ellipse(-15, -15, 30, 30);
        ill.push();
        ill.rotate(turtles[i].heading);
        ill.line(0, 0, 30, 0);
        ill.pop();
        ill.pop();
    }
    ill.pop();
    room.draw(ill, "web2")
    room.flush();
}, 250);


run();
