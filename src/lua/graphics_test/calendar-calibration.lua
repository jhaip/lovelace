require "zhelpers"
local room = require "helper"
local json = require "json"
local matrix = require "matrix"

SCREEN_WIDTH = 1280
SCREEN_HEIGHT = 720
CAMERA_IMAGE_WIDTH = 1280
CAMERA_IMAGE_HEIGHT = 720

function getSquareCalibrationList(w, h)
    return {{x=0, y=0}, {x=w, y=0}, {x=w, y=h}, {x=0, y=h}}
end

local SCREEN_SIZE = getSquareCalibrationList(SCREEN_WIDTH, SCREEN_HEIGHT)
local OFFSET = 200;

function getCalibrationRegionDefault()
    return {
        {x=OFFSET, y=OFFSET},
        {x=SCREEN_WIDTH - OFFSET, y=OFFSET},
        {x=SCREEN_WIDTH - OFFSET, y=SCREEN_HEIGHT - OFFSET},
        {x=OFFSET, y=SCREEN_HEIGHT - OFFSET}
    }
end

local SCREEN_SIZE_OFFSET_INNER = getCalibrationRegionDefault()

-- BASE_CALIBRATION should match the resolution of the camera?
local BASE_CALIBRATION = getSquareCalibrationList(SCREEN_WIDTH, SCREEN_HEIGHT)
-- calibrationRegion = getSquareCalibrationList(SCREEN_WIDTH, SCREEN_HEIGHT)

calibrationRegion = getCalibrationRegionDefault()
calendarRegion = {}
COMBINED_TRANSFORM = {}

function getPerspectiveTransform(src, dst)
    -- src table {{x=1, y=1}, ...}
    -- dst table
    -- order: Top left (TL), TR, BR, BL
    local a = matrix{
        {0, 0, 1, 0, 0, 0, 0, 0},
        {0, 0, 1, 0, 0, 0, 0, 0},
        {0, 0, 1, 0, 0, 0, 0, 0},
        {0, 0, 1, 0, 0, 0, 0, 0},
        {0, 0, 0, 0, 0, 1, 0, 0},
        {0, 0, 0, 0, 0, 1, 0, 0},
        {0, 0, 0, 0, 0, 1, 0, 0},
        {0, 0, 0, 0, 0, 1, 0, 0},
    }
    local b = matrix{{0}, {0}, {0}, {0}, {0}, {0}, {0}, {0}}
    for i = 1, 4 do
        a[i][1] = src[i].x
        a[i+4][4] = src[i].x
        a[i][2] = src[i].y
        a[i+4][5] = src[i].y
        a[i][7] = -src[i].x*dst[i].x
        a[i][8] = -src[i].y*dst[i].x
        a[i+4][7] = -src[i].x*dst[i].y
        a[i+4][8] = -src[i].y*dst[i].y
        b[i][1] = dst[i].x
        b[i+4][1] = dst[i].y
    end
    x = a:invert() * b
    return matrix{
      {x[1][1], x[2][1], x[3][1]},
      {x[4][1], x[5][1], x[6][1]},
      {x[7][1], x[8][1], 1},
    }
end

function projectPoint(homographyMatrix, pt)
    local r = homographyMatrix * matrix{{pt.x}, {pt.y}, {1}}
    return {x=r[1][1], y=r[2][1]}
end

function recalculateCombinedTransform()
    print("[[ Recalculating combined transform ]]")
    local calibrationTransformMatrix = getPerspectiveTransform(
        calibrationRegion,
        SCREEN_SIZE_OFFSET_INNER
    )
    local M = calibrationTransformMatrix
    local M2 = calibrationTransformMatrix
    if #calendarRegion > 0 then
        local projectedCalendarRegion = {
            projectPoint(calibrationTransformMatrix, calendarRegion[1]),
            projectPoint(calibrationTransformMatrix, calendarRegion[2]),
            projectPoint(calibrationTransformMatrix, calendarRegion[3]),
            projectPoint(calibrationTransformMatrix, calendarRegion[4]),
        }
        local screenToCalendarTransformMatrix = getPerspectiveTransform(
            SCREEN_SIZE,
            projectedCalendarRegion
        )
        M2 = screenToCalendarTransformMatrix
    end
    room.claim({
        {type="retract", fact={
            {"id", room.get_my_id_str()},
            {"id", "0"},
            {"postfix", ""},
        }},
        {type="claim", fact={
            {"id", room.get_my_id_str()},
            {"id", "0"},
            {"text", "camera"},
            {"integer", "1997"},
            {"text", "has"},
            {"text", "projector"},
            {"text", "calibration"},
            {"text", "TL"},
            {"text", "("},
            {"integer", tostring(calibrationRegion[1].x)},
            {"text", ","},
            {"integer", tostring(calibrationRegion[1].y)},
            {"text", ")"},
            {"text", "TR"},
            {"text", "("},
            {"integer", tostring(calibrationRegion[2].x)},
            {"text", ","},
            {"integer", tostring(calibrationRegion[2].y)},
            {"text", ")"},
            {"text", "BR"},
            {"text", "("},
            {"integer", tostring(calibrationRegion[3].x)},
            {"text", ","},
            {"integer", tostring(calibrationRegion[3].y)},
            {"text", ")"},
            {"text", "BL"},
            {"text", "("},
            {"integer", tostring(calibrationRegion[4].x)},
            {"text", ","},
            {"integer", tostring(calibrationRegion[4].y)},
            {"text", ")"},
            {"text", "@"},
            {"integer", tostring(os.time())},
        }},
        {type="claim", fact={
            {"id", room.get_my_id_str()},
            {"id", "0"},
            {"text", "camera"},
            {"text", "calibration"},
            {"text", "for"},
            {"integer", "1997"},
            {"text", "is"},
            {"float", tostring(M[1][1])},
            {"float", tostring(M[1][2])},
            {"float", tostring(M[1][3])},
            {"float", tostring(M[2][1])},
            {"float", tostring(M[2][2])},
            {"float", tostring(M[2][3])},
            {"float", tostring(M[3][1])},
            {"float", tostring(M[3][2])},
            {"float", tostring(M[3][3])},
        }},
        {type="claim", fact={
            {"id", room.get_my_id_str()},
            {"id", "0"},
            {"text", "calendar"},
            {"text", "calibration"},
            {"text", "for"},
            {"integer", "1997"},
            {"text", "is"},
            {"float", tostring(M2[1][1])},
            {"float", tostring(M2[1][2])},
            {"float", tostring(M2[1][3])},
            {"float", tostring(M2[2][1])},
            {"float", tostring(M2[2][2])},
            {"float", tostring(M2[2][3])},
            {"float", tostring(M2[3][1])},
            {"float", tostring(M2[3][2])},
            {"float", tostring(M2[3][3])},
        }}
    })
end

room.prehook(function()
    recalculateCombinedTransform()
end)

room.on({"$ $ region $id at $x1 $y1 $x2 $y2 $x3 $y3 $x4 $y4",
         "$ $ region $id has name calibration"}, function(results)
    calibrationRegion = getCalibrationRegionDefault()
    for i = 1, #results do
        local r = results[i]
        calibrationRegion = {
            {x=r.x1, y=r.y1},
            {x=r.x2, y=r.y2},
            {x=r.x3, y=r.y3},
            {x=r.x4, y=r.y4}
        }
    end
    recalculateCombinedTransform()
end)

room.on({"$ $ region $id at $x1 $y1 $x2 $y2 $x3 $y3 $x4 $y4",
         "$ $ region $id has name calendar"}, function(results)
    calendarRegion = {}
    for i = 1, #results do
        local r = results[i]
        calendarRegion = {
            {x=r.x1, y=r.y1},
            {x=r.x2, y=r.y2},
            {x=r.x3, y=r.y3},
            {x=r.x4, y=r.y4}
        }
    end
    recalculateCombinedTransform()
end)

local MY_ID_STR = "1996"
if #arg >= 1 then
    MY_ID_STR = arg[1]
    print("Set MY_ID_STR to:")
    print(MY_ID_STR)
end

room.init(false, MY_ID_STR)
