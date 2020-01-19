const { room, myId, run } = require('../helper2')(__filename);

const CANVAS_WIDTH = 1920;
const CANVAS_HEIGHT = 1080;

function getWeekNum(d) {
    let monthDayNumber = d.getDate();
    let firstDay = new Date(d.getFullYear(), d.getMonth(), 1);
    let firstDayDayOfWeek = firstDay.getDay();
    let weekNum = Math.floor((monthDayNumber + firstDayDayOfWeek - 1) / 7);
    return weekNum;
}

room.on(`wish calendar day $dateString was highlighted with color $color`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results) {
            results.forEach(({ dateString, color }) => {
                let date = new Date(Date.parse(dateString))
                let dateWithoutTimezone = new Date(date.getTime() + date.getTimezoneOffset() * 60 * 1000)
                let dayOfWeek = dateWithoutTimezone.getDay()
                let weekOfMonth = getWeekNum(dateWithoutTimezone)
                let ill = room.newIllumination()
                ill.nostroke();
                ill.fill(color)
                ill.rect(
                    dayOfWeek * Math.floor(CANVAS_WIDTH / 7),
                    weekOfMonth * Math.floor(CANVAS_HEIGHT / 5),
                    Math.floor(CANVAS_WIDTH / 7),
                    Math.floor(CANVAS_HEIGHT / 5)
                )
                room.draw(ill, "miniprojector")

            });
        }
        room.subscriptionPostfix();
    })


run();