const fs = require('fs');
const path = require('path');
const zmq = require('zeromq');
const uuidV4 = require('uuid/v4');

const randomId = () => 
    uuidV4();

function init(filename) {
    const scriptName = path.basename(filename);
    const scriptNameNoExtension = path.parse(scriptName).name;
    const logPath = filename.replace(scriptName, 'logs/' + scriptNameNoExtension + '.log')
    const access = fs.createWriteStream(logPath)
    // process.stdout.write = process.stderr.write = access.write.bind(access);
    // process.on('uncaughtException', function (err) {
    //     console.error((err && err.stack) ? err.stack : err);
    // })
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

    const room = {
        subscribe: (...args) => {
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
        on: (...args) => {
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
        assert: fact => {
            publisher.send(`....CLAIM${MY_ID_STR}${fact}`);
        },
        retract: query => {
            publisher.send(`..RETRACT${MY_ID_STR}${query}`);
        }
    }

    subscriber.on('message', (request) => {
        console.log("GOT MESSAGE")
        const msg = request.toString();
        console.log(msg)
        const source_len = 4
        const SUBSCRIPTION_ID_LEN = (randomId()).length
        const id = msg.slice(source_len, source_len + SUBSCRIPTION_ID_LEN)
        console.log(`ID: ${id} == ${init_ping_id}`)
        const val = msg.slice(source_len + SUBSCRIPTION_ID_LEN)
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
            callback(val)
        }
    });

    setTimeout(() => {
        publisher.send(`.....PING${MY_ID_STR}${init_ping_id}`)
        console.log("SEND PING!")
    }, 1000)

    return {
        room, myId, scriptName
    }
}

module.exports = init
