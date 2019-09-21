const fs = require('fs');
const path = require('path');
const zmq = require('zeromq');
const uuidV4 = require('uuid/v4');
const child_process = require("child_process");
var initTracer = require('jaeger-client').initTracer;
var opentracing = require('opentracing');
var flatbuffers = require('./flatbuffers').flatbuffers;
var roomupdatefbs = require('./request_generated').roomupdatefbs; // Generated by `flatc`.

// See schema https://github.com/jaegertracing/jaeger-client-node/blob/master/src/configuration.js#L37
var config = {
    serviceName: 'room-service',
    'reporter': {
        'logSpans': true,
        'agentHost': 'localhost',
        'agentPort': 6832
    },
    'sampler': {
        'type': 'const',
        'param': 1.0
    }
};
var options = {
    // metrics: metrics,
    // logger: logger,
};
// var tracer = initTracer(config, options);  // uncomment to use real tracer
var tracer = new opentracing.Tracer();  // no-op dummy tracer

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

function stringToUint(string) {
    var string = unescape(encodeURIComponent(string)),
        charList = string.split(''),
        uintArray = [];
    for (var i = 0; i < charList.length; i++) {
        uintArray.push(charList[i].charCodeAt(0));
    }
    return new Uint8Array(uintArray);
}

function uintToString(uintArray) {
    var encodedString = String.fromCharCode.apply(null, uintArray),
        decodedString = decodeURIComponent(escape(encodedString));
    return decodedString;
}

function makePingMessage(source, pingId) {
    var builder = new flatbuffers.Builder(1024);

    var facts = roomupdatefbs.RoomUpdate.createFactsVector(builder, [])

    var updateSource = builder.createString(source)
    var updateSubId = builder.createString(pingId)

    roomupdatefbs.RoomUpdate.startRoomUpdate(builder)
    roomupdatefbs.RoomUpdate.addType(builder, roomupdatefbs.UpdateType.Ping)
    roomupdatefbs.RoomUpdate.addSource(builder, updateSource)
    roomupdatefbs.RoomUpdate.addSubscriptionId(builder, updateSubId)
    roomupdatefbs.RoomUpdate.addFacts(builder, facts)
    var update = roomupdatefbs.RoomUpdate.endRoomUpdate(builder)

    // Ping only has 1 RoomUpdate
    var updates = roomupdatefbs.RoomUpdates.createUpdatesVector(builder, [update])
    roomupdatefbs.RoomUpdates.startRoomUpdates(builder)
    roomupdatefbs.RoomUpdates.addUpdates(builder, updates)
    var full_updates_msg = roomupdatefbs.RoomUpdates.endRoomUpdates(builder)

    builder.finish(full_updates_msg)
    var msg_buf = builder.asUint8Array(); // Of type `Uint8Array`.
    return Buffer.from(msg_buf)
}

function makeSubscriptionMessage(source, subscriptionId, subscriptionQueryParts) {
    var builder = new flatbuffers.Builder(1024);

    var subscriptionFactArray = new Array(subscriptionQueryParts.length);
    for (let i = 0; i < subscriptionQueryParts.length; i += 1) {
        var factValue = roomupdatefbs.Fact.createValueVector(
            builder,
            stringToUint(subscriptionQueryParts[i])
        )
        roomupdatefbs.Fact.startFact(builder)
        roomupdatefbs.Fact.addType(builder, roomupdatefbs.FactType.Text)
        roomupdatefbs.Fact.addValue(builder, factValue)
        subscriptionFactArray[i] = roomupdatefbs.Fact.endFact(builder)
    }
    var facts = roomupdatefbs.RoomUpdate.createFactsVector(builder, subscriptionFactArray)

    var updateSource = builder.createString(source)
    var updateSubId = builder.createString(subscriptionId)

    roomupdatefbs.RoomUpdate.startRoomUpdate(builder)
    roomupdatefbs.RoomUpdate.addType(builder, roomupdatefbs.UpdateType.Subscribe)
    roomupdatefbs.RoomUpdate.addSource(builder, updateSource)
    roomupdatefbs.RoomUpdate.addSubscriptionId(builder, updateSubId)
    roomupdatefbs.RoomUpdate.addFacts(builder, facts)
    var update = roomupdatefbs.RoomUpdate.endRoomUpdate(builder)

    // Subscription only has 1 RoomUpdate
    var updates = roomupdatefbs.RoomUpdates.createUpdatesVector(builder, [update])
    roomupdatefbs.RoomUpdates.startRoomUpdates(builder)
    roomupdatefbs.RoomUpdates.addUpdates(builder, updates)
    var full_updates_msg = roomupdatefbs.RoomUpdates.endRoomUpdates(builder)

    builder.finish(full_updates_msg)
    var msg_buf = builder.asUint8Array(); // Of type `Uint8Array`.
    return Buffer.from(msg_buf)
}

function checkEndian() {
    var arrayBuffer = new ArrayBuffer(2);
    var uint8Array = new Uint8Array(arrayBuffer);
    var uint16array = new Uint16Array(arrayBuffer);
    uint8Array[0] = 0xAA; // set first byte
    uint8Array[1] = 0xBB; // set second byte
    if (uint16array[0] === 0xBBAA) return "little endian";
    if (uint16array[0] === 0xAABB) return "big endian";
    else throw new Error("Something crazy just happened");
}

function makeBatchMessage(source, batched_calls) {
    var builder = new flatbuffers.Builder(1024*8);
    const batchMessageTypeToMessageTypeEnum = {
        "claim": roomupdatefbs.UpdateType.Claim,
        "retract": roomupdatefbs.UpdateType.Retract,
        "death": roomupdatefbs.UpdateType.Death
    }
    const factTypeStringToTypeEnum = {
        "id": roomupdatefbs.FactType.Id,
        "text": roomupdatefbs.FactType.Text,
        "integer": roomupdatefbs.FactType.Integer,
        "float": roomupdatefbs.FactType.Float,
        "binary": roomupdatefbs.FactType.Binary
    }

    var batchedUpdatesArray = new Array(batched_calls.length);
    for (let i = 0; i < batched_calls.length; i += 1) {
        var factArray = new Array(batched_calls[i].fact.length);
        for (let k = 0; k < factArray.length; k += 1) {
            let factPart = batched_calls[i].fact[k];
            let factType = factTypeStringToTypeEnum[factPart[0]];
            // should the value value be encoded differently depending on the type?
            let factValue;
            if (
                factType == roomupdatefbs.FactType.Integer ||
                factType == roomupdatefbs.FactType.Float
            ) {
                // encode integer or float
                // using little endian
                // buf.writeUInt32LE(+factPart[1])
                // TODO: handle float
                // const buf = Buffer.allocUnsafe(4);
                // console.log("writing int")
                // console.log(factPart[1])
                // buf.writeUInt32LE(+factPart[1])
                // factValue = roomupdatefbs.Fact.createValueVector(
                //     builder,
                //     buf.readUInt8(0)
                // )
                factValue = roomupdatefbs.Fact.createValueVector(
                    builder,
                    stringToUint(`${factPart[1]}`)
                )
            } else if (factType == roomupdatefbs.FactType.Binary) {
                factValue = roomupdatefbs.Fact.createValueVector(
                    builder,
                    factPart[1]
                )
            } else {
                factValue = roomupdatefbs.Fact.createValueVector(
                    builder,
                    stringToUint(factPart[1])
                )
            }

            roomupdatefbs.Fact.startFact(builder)
            roomupdatefbs.Fact.addType(builder, factType)
            roomupdatefbs.Fact.addValue(builder, factValue)
            factArray[k] = roomupdatefbs.Fact.endFact(builder)
        }
        var facts = roomupdatefbs.RoomUpdate.createFactsVector(builder, factArray)

        var updateType = batchMessageTypeToMessageTypeEnum[batched_calls[i].type];
        var updateSource = builder.createString(source)
        var updateSubId = builder.createString("") // batch calls don't have subscription IDs

        roomupdatefbs.RoomUpdate.startRoomUpdate(builder)
        roomupdatefbs.RoomUpdate.addType(builder, updateType)
        roomupdatefbs.RoomUpdate.addSource(builder, updateSource)
        roomupdatefbs.RoomUpdate.addSubscriptionId(builder, updateSubId)
        roomupdatefbs.RoomUpdate.addFacts(builder, facts)
        batchedUpdatesArray[i] = roomupdatefbs.RoomUpdate.endRoomUpdate(builder)
    }

    // Subscription only has 1 RoomUpdate
    var updates = roomupdatefbs.RoomUpdates.createUpdatesVector(builder, batchedUpdatesArray)
    roomupdatefbs.RoomUpdates.startRoomUpdates(builder)
    roomupdatefbs.RoomUpdates.addUpdates(builder, updates)
    var full_updates_msg = roomupdatefbs.RoomUpdates.endRoomUpdates(builder)

    builder.finish(full_updates_msg)
    var msg_buf = builder.asUint8Array(); // Of type `Uint8Array`.
    // console.log("BATCH MESSAGE:")
    // console.log(msg_buf.length)
    // console.log(msg_buf.join(' '))
    return Buffer.from(msg_buf)
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

    const rpc_url = "localhost";
    const MY_ID_STR = getIdStringFromId(myId);
    client = zmq.socket('dealer');
    client.identity = MY_ID_STR;
    client.connect(`tcp://${rpc_url}:5570`);

    let init_ping_id = randomId()
    let subscription_ids = {}
    let server_listening = false
    let sent_ping = false;
    let batched_calls = []
    const DEFAULT_SUBSCRIPTION_ID = 0;
    let currentSubscriptionId = DEFAULT_SUBSCRIPTION_ID;
    var wireCtx;

    const sleep = (ms) => {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    const waitForServerListening = () => {
        // Sends a single ping and loops until server_listening is set to true
        // server_listening is set to true outside this function in the message received callback
        return new Promise(async resolve => {
            if (server_listening === false) {
                if (sent_ping == false) {
                    // var s2 = new Uint8Array(2)
                    // s2[0] = 9
                    // s2[1] = 254
                    // client.send(Buffer.from(s2))
                    client.send(makePingMessage(MY_ID_STR, init_ping_id))
                    sent_ping = true;
                }
                while (server_listening === false) {
                    await sleep(100);
                    // await null; // prevents app from hanging
                }
                // console.log(`wait for server listening done ${MY_ID_STR}`)
            }
            resolve();
        });
    }

    const room = {
        setCtx: ctx => {
            wireCtx = ctx;
            this.wireCtx = ctx;
        },
        wireCtx: () => {
            return this.wireCtx;
        },
        onRaw: async (...args) => {
            await waitForServerListening();
            const query_strings = args.slice(0, -1)
            const callback = args[args.length - 1]
            const subscription_id = randomId()
            const subscriptionMsg = makeSubscriptionMessage(MY_ID_STR, subscription_id, query_strings)
            subscription_ids[subscription_id] = callback
            client.send(subscriptionMsg);
        },
        onGetSource: async (...args) => {
            const sourceVariableName = args[0]
            const query_strings = args.slice(1, -1).map(s => `$${sourceVariableName} $ ${s}`)
            const callback = args[args.length - 1]
            room.onRaw(...query_strings, callback)
        },
        on: async (...args) => {
            const query_strings = args.slice(0, -1).map(s => `$ $ ${s}`)
            const callback = args[args.length - 1]
            room.onRaw(...query_strings, callback)
        },
        assertNow: (fact) => {
            client.send([`....CLAIM${MY_ID_STR}${fact}`]);
        },
        assertForOtherSource: (otherSource, fact) => {
            batched_calls.push({
                "type": "claim",
                "fact": [
                    ["id", otherSource],
                    ["id", `${DEFAULT_SUBSCRIPTION_ID}`]
                ].concat(fullyParseFact(fact))
            })
        },
        assert: (...args) => {
            batched_calls.push({
                "type": "claim",
                "fact": [
                    ["id", MY_ID_STR],
                    ["id", `${currentSubscriptionId}`]
                ].concat(fullyParseFact(args))
            })
        },
        retractNow: (query) => {
            client.send([`..RETRACT${MY_ID_STR}${query}`]);
        },
        retractRaw: (...args) => {
            // TODO: need to push into an array specific to the subsciber, in case there are multiple subscribers in one client
            batched_calls.push({ "type": "retract", "fact": fullyParseFact(args) })
        },
        retractMine: (...args) => {
            if (typeof args === "string") {
                room.retractRaw(`#${MY_ID_STR} $ ${args}`)
            } else if (Array.isArray(args)) {
                room.retractRaw(...[["id", MY_ID_STR], ["variable", ""]].concat(args))
            }
        },
        retractMineFromThisSubscription: (...args) => {
            if (typeof args === "string") {
                room.retractRaw(`#${MY_ID_STR} #${currentSubscriptionId} ${args}`)
            } else if (Array.isArray(args)) {
                room.retractRaw(...[["id", MY_ID_STR], ["id", `${currentSubscriptionId}`]].concat(args))
            }
        },
        retractFromSource: (...args) => {
            const source = args[0]
            const retractArgs = args.slice(1);
            if (typeof retractArgs === "string") {
                room.retractRaw(`#${source} $ ${retractArgs}`)
            } else if (Array.isArray(retractArgs)) {
                room.retractRaw(...[["id", `${source}`], ["variable", ""]].concat(retractArgs))
            }
        },
        retractAll: (...args) => {
            if (typeof args === "string") {
                room.retractRaw(`$ $ ${args}`)
            } else if (Array.isArray(args)) {
                room.retractRaw(...[["variable", ""], ["variable", ""]].concat(args))
            }
        },
        flush: () => {
            // TODO: need to push into an array specific to the subsciber, in case there are multiple subscribers in one client
            room.batch(batched_calls);
            batched_calls = [];
        },
        batch: batched_calls => {
            client.send(makeBatchMessage(MY_ID_STR, batched_calls));
        },
        cleanup: () => {
            room.retractMine(`%`)
        },
        cleanupOtherSource: (otherSource) => {
            room.batch([{ "type": "death", "fact": [["id", otherSource]] }])
        },
        draw: (illumination, target) => {
            target = typeof target === 'undefined' ? myId : target;
            room.assert(`draw graphics`, ["text", illumination.toString()], `on ${target}`)
        },
        newIllumination: () => {
            return new Illumination()
        },
        subscriptionPrefix: id => {
            currentSubscriptionId = id;
            room.retractMineFromThisSubscription(["postfix", ""])
        },
        subscriptionPostfix: () => {
            currentSubscriptionId = 0;
        }
    }

    function deserializeRoomResponseMessage(data) {
        /*
        Return something like:
        {
            "source": ...
            "subscriptionId": ...
            "results": [{"x": 5, "y": 3.002, "z": "Hello"}, {"x": 3, "y": 0.0, "z": "Two"}]
        }
        */
        if (Buffer.isBuffer(data)) {
            data = new Uint8Array(data)
        }
        var buf = new flatbuffers.ByteBuffer(data);
        var room_response_obj = roomupdatefbs.RoomResponse.getRootAsRoomResponse(buf)
        let returnObj = {
            "source": room_response_obj.source(),
            "subscriptionId": room_response_obj.subscriptionId(),
            "results": new Array(room_response_obj.resultSetsLength())
        }
        for (let i = 0; i < returnObj.results.length; i += 1) {
            let returnObjResult = {};
            var result_set = room_response_obj.resultSets(i)
            for (let k = 0; k < result_set.resultsLength(); k += 1) {
                var result = result_set.results(k)
                if (
                    result.type() === roomupdatefbs.FactType.Integer ||
                    result.type() === roomupdatefbs.FactType.Float
                ) {
                    returnObjResult[result.variableName()] = +uintToString(result.valueArray())
                } else if (result.type() === roomupdatefbs.FactType.Binary) {
                    returnObjResult[result.variableName()] = result.valueArray()
                } else {
                    returnObjResult[result.variableName()] = uintToString(result.valueArray())
                }
            }
            returnObj.results[i] = returnObjResult;
        }
        return returnObj;
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

    client.on('message', (request) => {
        const span = tracer.startSpan(`client-${myId}-recv`, { childOf: room.wireCtx() });
        // console.log("RAW RECV:");
        // console.log(request); // a buffer
        const roomResponse = deserializeRoomResponseMessage(request);
        // console.log(roomResponse);
        const id = roomResponse.subscriptionId;
        if (id == init_ping_id) {
            const val = roomResponse.results;
            server_listening = true
            console.log(`SERVER LISTENING!! ${MY_ID_STR} ${val}`)
            room.setCtx(tracer.extract(opentracing.FORMAT_TEXT_MAP, {"uber-trace-id": val}));
        } else if (id in subscription_ids) {
            callback = subscription_ids[id]
            // room.cleanup()
            // const callbackSpan = tracer.startSpan(`client-${myId}-callbackrecv`, { childOf: span });
            callback(roomResponse.results)
            // callbackSpan.finish();
            room.flush()
        } else {
            console.log("unknown subscription ID...")
        }
        span.finish();
    });

    const run = async () => {
        await waitForServerListening();
        room.flush()
    }

    const afterServerConnects = async (callback) => {
        await waitForServerListening();
        callback();
    }

    return {
        room, myId, scriptName, MY_ID_STR, run, getIdFromProcessName, getIdStringFromId, tracer, afterServerConnects
    }
}

module.exports = init
