const fs = require('fs');
const path = require('path');
const zmq = require('zeromq');
const uuidV4 = require('uuid/v4');
const child_process = require("child_process");

const randomId = () => 
    uuidV4();

function sleep(millis) {
    return new Promise(resolve => setTimeout(resolve, millis));
}

const stringToTerm = x => {
    if (x[0] === `"`) {
        return ["text", x.slice(1, -1)]
    }
    if (isNaN(x) || x === "") {
        if (x[0] === "#") {
            return ["id", x.slice(1)]
        }
        if (x[0] === "$") {
            return ["variable", x.slice(1)]
        }
        if (x[0] === "%") {
            return ["postfix", x.slice(1)]
        }
        return ["text", x]
    }
    if (x.indexOf(".") === -1) {
        return ["integer", (+x).toString()]
    }
    return ["float", (+x).toFixed(6)]
}

const tokenizeString = str => {
    // from https://stackoverflow.com/questions/2817646/javascript-split-string-on-space-or-on-quotes-to-array
    var spacedStr = str.trim().replace(/\)/g, ' ) ').replace(/\(/g, ' ( ').replace(/,/g, ' , ').trim();
    var aStr = spacedStr.match(/[\w,()@#:\-\.\$\%]+|"[^"]+"/g), i = aStr.length;
    while(i--){
        aStr[i] = aStr[i].replace(/"/g,"");
    }
    return aStr
}

const fullyParseFact = q => {    
    if (typeof q === "string") {
        const q_tokens = tokenizeString(q)
        return q_tokens.map(x => stringToTerm(x))
    } else if (Array.isArray(q)) {
        let terms = [];
        for (var i = 0, len = q.length; i < len; i++) {
            if (typeof q[i] === "string") {
                const q_tokens = tokenizeString(q[i])
                terms = terms.concat(q_tokens.map(x => stringToTerm(x)))
            } else {
                terms = terms.concat([q[i]])
            }
        }
        return terms
    }
}

function getIdFromProcessName(scriptName) {
    return (scriptName.split(".")[0]).split("__")[0]
}

function getIdStringFromId(id) {
    return String(id).padStart(4, '0')
}

class Illumination {
    constructor() {
        this.illuminations = [];
        this.add = (type, opts) => {
            this.illuminations.push({ "type": type, "options": opts })
        }
        this.addColorType = (type, opts) => {
            opts = (opts.length === 1) ? opts[0] : opts;
            this.add(type, opts);
        }
    }
    rect(x, y, w, h) { this.add("rectangle", {"x": x, "y": y, "w": w, "h": h}) }
    ellipse(x, y, w, h) { this.add("ellipse", { "x": x, "y": y, "w": w, "h": h }) }
    text(x, y, txt) { this.add("text", { "x": x, "y": y, "text": txt }) }
    line(x1, y1, x2, y2) { this.add("line", [x1, y1, x2, y2]) }
    // point format: [[x1, y1], [x2, y2], ...]
    polygon(points) { this.add("polygon", points) }
    // color format: string, [r, g, b], or [r, g, b, a]
    fill(...color) { this.addColorType("fill", color) }
    stroke(...color) { this.addColorType("stroke", color) }
    nostroke() { this.add("nostroke", []) }
    nofill() { this.add("nofill", []) }
    strokewidth(width) { this.add("strokewidth", width) }
    fontsize(width) { this.add("fontsize", width) }
    fontcolor(...color) { this.addColorType("fontcolor", color) }
    push() { this.add("push", []) }
    pop() { this.add("pop", []) }
    translate(x, y) { this.add("translate", { "x": x, "y": y }) }
    rotate(radians) { this.add("rotate", radians) }
    scale(x, y) { this.add("scale", { "x": x, "y": y }) }
    toString() { return JSON.stringify(this.illuminations) }
}

function init(filename) {
    const scriptName = path.basename(filename);
    const scriptNameNoExtension = path.parse(scriptName).name;
    const logPath = filename.replace(scriptName, 'logs/' + scriptNameNoExtension + '.log')
    const access = fs.createWriteStream(logPath)
    process.stdout.write = process.stderr.write = access.write.bind(access);
    process.on('uncaughtException', function (err) {
        console.error((err && err.stack) ? err.stack : err);
    })
    const myId = getIdFromProcessName(scriptName);

    const subscriber = zmq.socket('sub');
    const publisher = zmq.socket('pub');
    const rpc_url = "localhost";
    subscriber.connect(`tcp://${rpc_url}:5555`)
    publisher.connect(`tcp://${rpc_url}:5556`)
    const MY_ID_STR = getIdStringFromId(myId);
    subscriber.subscribe(MY_ID_STR);

    let init_ping_id = randomId()
    let select_ids = {}
    let subscription_ids = {}
    let server_listening = false
    let batched_calls = []

    const waitForServerListening = () => {
        return new Promise(async resolve => {
            while (server_listening === false) {
                publisher.send(`.....PING${MY_ID_STR}${init_ping_id}`)
                // console.log("SEND PING!")
                await sleep(500)
            }
            resolve();
        });
    }

    const room = {
        onRaw: async (...args) => {
            console.log("pre wait for server")
            await waitForServerListening();
            console.log("post wait for server")
            const query_strings = args.slice(0, -1)
            const callback = args[args.length - 1]
            const subscription_id = randomId()
            const query_msg = {
                "id": subscription_id,
                "facts": query_strings
            }
            const query_msg_str = JSON.stringify(query_msg)
            subscription_ids[subscription_id] = callback
            publisher.send(`SUBSCRIBE${MY_ID_STR}${query_msg_str}`);
            console.log("send ON listen")
        },
        onGetSource: async (...args) => {
            const sourceVariableName = args[0]
            const query_strings = args.slice(1, -1).map(s => `$${sourceVariableName} ${s}`)
            const callback = args[args.length - 1]
            onRaw(...query_strings, callback)
        },
        on: async (...args) => {
            const query_strings = args.slice(1, -1).map(s => `$ ${s}`)
            const callback = args[args.length - 1]
            onRaw(...query_strings, callback)
        },
        select: async (...args) => {
            await waitForServerListening();
            const query_strings = args.slice(0, -1)
            const callback = args[args.length - 1]
            const select_id = randomId()
            const query_msg = {
                "id": select_id,
                "facts": query_strings
            }
            const query_msg_str = JSON.stringify(query_msg)
            select_ids[select_id] = callback
            publisher.send(`...SELECT${MY_ID_STR}${query_msg_str}`);
            console.log("SEND!")
        },
        assertNow: (fact) => {
            publisher.send(`....CLAIM${MY_ID_STR}${fact}`);
        },
        assertForOtherSource: (otherSource, fact) => {
            console.error("assertForOtherSource")
            batched_calls.push({ "type": "claim", "fact": [["id", otherSource]].concat(fullyParseFact(fact)) })
        },
        assert: (...args) => {
            // TODO: need to push into an array specific to the subsciber, in case there are multiple subscribers in one client
            batched_calls.push({ "type": "claim", "fact": [["id", MY_ID_STR]].concat(fullyParseFact(args)) })
        },
        retractNow: (query) => {
            publisher.send(`..RETRACT${MY_ID_STR}${query}`);
        },
        retractRaw: (...args) => {
            // TODO: need to push into an array specific to the subsciber, in case there are multiple subscribers in one client
            batched_calls.push({ "type": "retract", "fact": fullyParseFact(args) })
        },
        retractMine: (...args) => {
            retractRaw(args.map(a => {
                if (typeof a === "string") {
                    return `#${MY_ID_STR} ${a}`
                } else if (Array.isArray(a)) {
                    return [["id", MY_ID_STR]].concat(a)
                }
            }))
        },
        retractFromSource: (...args) => {
            const source = args[0]
            retractRaw(args.slice(1, -1).map(a => {
                if (typeof a === "string") {
                    return `#${source} ${a}`
                } else if (Array.isArray(a)) {
                    return [["id", source]].concat(a)
                }
            }))
        },
        retractAll: (...args) => {
            retractRaw(args.map(a => {
                if (typeof a === "string") {
                    return `$ ${a}`
                } else if (Array.isArray(a)) {
                    return [["variable", "$"]].concat(a)
                }
            }))
        },
        flush: () => {
            // TODO: need to push into an array specific to the subsciber, in case there are multiple subscribers in one client
            room.batch(batched_calls);
            batched_calls = [];
        },
        batch: batched_calls => {
            // console.log("SEINDING BATCH", batched_calls)
            const fact_str = JSON.stringify(batched_calls)
            publisher.send(`....BATCH${MY_ID_STR}${fact_str}`);
        },
        cleanup: () => {
            room.retract(`#${MY_ID_STR} %`)
        },
        cleanupOtherSource: (otherSource) => {
            const fact_str = JSON.stringify([{ "type": "death", "fact": [["id", otherSource]] }])
            publisher.send(`....BATCH${MY_ID_STR}${fact_str}`);
        },
        draw: (illumination, target) => {
            target = typeof target === 'undefined' ? myId : target;
            room.assert(`draw graphics`, ["text", illumination.toString()], `on ${target}`)
        },
        newIllumination: () => {
            return new Illumination()
        },
    }

    const parseResult = result_str => {
        return JSON.parse(result_str).map(result => {
            let newResult = {}
            for (let key in result) {
                // result[key] = ["type", "value"]
                // for legacy reasons, we just want the value
                const value_type = result[key][0]
                if (value_type === "integer" || value_type === "float") {
                    newResult[key] = +result[key][1]    
                } else {
                    newResult[key] = result[key][1]
                }
            }
            return newResult
        })
    }

    subscriber.on('message', (request) => {
        console.log("GOT MESSAGE")
        const msg = request.toString();
        console.log(msg)
        const source_len = 4
        const SUBSCRIPTION_ID_LEN = (randomId()).length
        const SERVER_SEND_TIME_LEN = 13
        const id = msg.slice(source_len, source_len + SUBSCRIPTION_ID_LEN)
        console.log(`ID: ${id} == ${init_ping_id}`)
        const val = msg.slice(source_len + SUBSCRIPTION_ID_LEN + SERVER_SEND_TIME_LEN)
        if (id == init_ping_id) {
            server_listening = true
            console.log("SERVER LISTENING!!")
            return
        }
        if (id in select_ids) {
            const callback = select_ids[id]
            delete select_ids[id]
            callback(val)
        } else if (id in subscription_ids) {
            callback = subscription_ids[id]
            // room.cleanup()
            // console.log("found match")
            const r = parseResult(val)
            // console.log(r)
            callback(r)
            // console.log("flushing")
            room.flush()
        } else {
            console.log("unknown subscription ID...")
        }
    });

    const run = async () => {
        await waitForServerListening();
        room.flush()
    }

    return {
        room, myId, scriptName, MY_ID_STR, run, getIdFromProcessName, getIdStringFromId
    }
}

module.exports = init
