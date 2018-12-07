const { room, myId, run } = require('../helper2')(__filename);

var FACTS = {};
var SUBSCRIPTIONS = {};
var nonBootFactSubcriptions = {};
const FONT_SIZE = 16;
const ORIGIN = [0, 0];

function render() {
    console.error(FACTS);
    console.error(SUBSCRIPTIONS);
    room.cleanup();
    let ill = room.newIllumination();
    ill.fontsize(FONT_SIZE);
    let offset = 0;
    ill.fontcolor(255, 0, 0);
    for (var key in SUBSCRIPTIONS) {
        let sortedSubscriptions = SUBSCRIPTIONS[key].map(({ subscription }) => subscription).sort();
        sortedSubscriptions.forEach(subscription => {
            ill.text(ORIGIN[0], (ORIGIN[1] + (offset) * FONT_SIZE * 1.3), `#${key} ${subscription}`);
            offset += 1;
        });
    }
    ill.fontcolor(0, 255, 0);
    for (var key in FACTS) {
        let sortedFacts = FACTS[key].map(({ fact }) => fact).sort();
        sortedFacts.forEach(fact => {
            ill.text(ORIGIN[0], (ORIGIN[1] + (offset) * FONT_SIZE * 1.3), `#${key} ${fact}`);
            offset += 1;
        });
    }
    room.draw(ill);
}

room.on(
    `$ camera 1 sees paper $id at TL ( $ , $ ) TR ( $ , $ ) BR ( $ , $ ) BL ( $ , $ ) @ $time`,
    results => {
        console.error("seeing non-boot papers:")
        console.error(results)
        for (let i=0; i<results.length; i+=1) {
            const id = results[i].id.toString().padStart(4, '0');
            if (nonBootFactSubcriptions[id] !== "SUBSCRIBED") {
                nonBootFactSubcriptions[id] = "SUBSCRIBED";
                room.on(`#${id} %fact`, factSubscriptionResult => {
                    FACTS[id] = factSubscriptionResult;
                    render();
                });
                room.on(`subscription #${id} %subscription`, subscriptionSubscriptionResult => {
                    SUBSCRIPTIONS[id] = subscriptionSubscriptionResult;
                    render();
                });
            }
        }
    }
)