const { room, myId, run } = require('../helper2')(__filename);

// Banana emoji (JPG) as a b64 string.
b64_img_str = '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'

let ill = room.newIllumination()
ill.image(0, 0, 50, 50, b64_img_str)
room.draw(ill)

run();