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
            return ["source", x.slice(1)]
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

const fullyParseFact = q => {
    const tokenizeString = s => s.trim().replace(/\)/g, ' ) ').replace(/\(/g, ' ( ').replace(/,/g, ' , ').split(/\s+/)
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

function init(filename) {
    const scriptName = path.basename(filename);
    const scriptNameNoExtension = path.parse(scriptName).name;
    const logPath = filename.replace(scriptName, 'logs/' + scriptNameNoExtension + '.log')
    const access = fs.createWriteStream(logPath)
    process.stdout.write = process.stderr.write = access.write.bind(access);
    process.on('uncaughtException', function (err) {
        console.error((err && err.stack) ? err.stack : err);
    })
    const myId = (scriptName.split(".")[0]).split("__")[0]

    const subscriber = zmq.socket('sub');
    const publisher = zmq.socket('pub');
    const rpc_url = "localhost";
    subscriber.connect(`tcp://${rpc_url}:5555`)
    publisher.connect(`tcp://${rpc_url}:5556`)
    const MY_ID_STR = String(myId).padStart(4, '0');
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
        subscribe: async (...args) => {
            await waitForServerListening();
            const query_strings = args.slice(0, -1)
            const callback = args[args.length - 1]
            const subscription_id = randomId()
            const query_msg = {
                "id": subscription_id,
                "facts": query_strings
            }
            console.log("query_msg:")
            console.log(query_msg)
            const query_msg_str = JSON.stringify(query_msg)
            console.log(query_msg_str)
            subscription_ids[subscription_id] = callback
            publisher.send(`SUBSCRIBE${MY_ID_STR}${query_msg_str}`);
        },
        on: async (...args) => {
            await waitForServerListening();
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
        },
        select: query_strings => {
            const select_id = randomId()
            const query_msg = {
                "id": select_id,
                "facts": query_strings
            }
            const query_msg_str = JSON.stringify(query_msg)
            select_ids[select_id] = callback
            publisher.send(`...SELECT${MY_ID_STR}${query_msg_str}`);
        },
        assertNow: (fact) => {
            publisher.send(`....CLAIM${MY_ID_STR}${fact}`);
        },
        assertForOtherSource: (otherSource, fact) => {
            console.error("assertForOtherSource")
            batched_calls.push({ "type": "claim", "fact": [["source", otherSource]].concat(fullyParseFact(fact)) })
        },
        assert: (...args) => {
            // TODO: need to push into an array specific to the subsciber, in case there are multiple subscribers in one client
            batched_calls.push({ "type": "claim", "fact": [["source", MY_ID_STR]].concat(fullyParseFact(args)) })
        },
        retractNow: (query) => {
            publisher.send(`..RETRACT${MY_ID_STR}${query}`);
        },
        retract: (...args) => {
            // TODO: need to push into an array specific to the subsciber, in case there are multiple subscribers in one client
            batched_calls.push({ "type": "retract", "fact": fullyParseFact(args) })
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
        }
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
            callback(parseResult(val))
            room.flush()
        }
    });

    const run = async () => {
        await waitForServerListening();
        room.flush()
    }

    return {
        room, myId, scriptName, MY_ID_STR, run
    }
}

module.exports = init
