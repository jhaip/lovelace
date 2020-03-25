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
-- BASE_CALIBRATION should match the resolution of the camera?
local BASE_CALIBRATION = getSquareCalibrationList(SCREEN_WIDTH, SCREEN_HEIGHT)
calibrationRegion = getSquareCalibrationList(SCREEN_WIDTH, SCREEN_HEIGHT)
calendarRegion = getSquareCalibrationList(SCREEN_WIDTH, SCREEN_HEIGHT)
COMBINED_TRANSFORM = {}

graphics_cache = {}
font = false

local colors = {
    white={255, 255, 255},
    red={255, 0, 0},
    green={0, 255, 0},
    blue={0, 0, 255},
    black={0, 0, 0},
    yellow={255, 255, 0},
    purple={128, 0, 128},
    cyan={0, 255, 255},
    orange={255, 165, 0},
}

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

function projectPoint(homographyMatrix, x, y)
    local r = homographyMatrix * matrix{{x}, {y}, {1}}
    return {r[1][1], r[2][1]}
end

function convertFromMatrixToTransform(M)
    -- https://forum.openframeworks.cc/t/quad-warping-an-entire-opengl-view-solved/509/10
    local transform = love.math.newTransform()
    transform:setMatrix(
        M[1][1], M[1][2], 0, M[1][3],
        M[2][1], M[2][2], 0, M[2][3],
        0,       0,       1, 0,
        M[3][1], M[3][2], 0, 1
    )
    return transform
end

function recalculateCombinedTransform()
    print("[[ Recalculating combined transform ]]")
    local calendarTransformMatrix = getPerspectiveTransform(
        SCREEN_SIZE,
        calendarRegion
    )
    local calendarTransform = convertFromMatrixToTransform(calendarTransformMatrix)
    local calibrationTransformMatrix = getPerspectiveTransform(
        calibrationRegion,
        SCREEN_SIZE
    )
    local calibrationTransform = convertFromMatrixToTransform(calendarTransformMatrix)
    -- local combined_matrix = calendarTransformMatrix * calibrationTransformMatrix;
    -- COMBINED_TRANSFORM = convertFromMatrixToTransform(combined_matrix)
    COMBINED_TRANSFORM = calendarTransform:clone()
    COMBINED_TRANSFORM:apply(calibrationTransform)
end

recalculateCombinedTransform()

room.on({"$ $ region $id at $x1 $y1 $x2 $y2 $x3 $y3 $x4 $y4",
         "$ $ region $id has name calibration"}, function(results)
    for i = 1, #results do
        local r = results[i]
        calibrationRegion = {
            {x=r.x1, y=r.y1},
            {x=r.x2, y=r.y2},
            {x=r.x3, y=r.y3},
            {x=r.x4, y=r.y4}
        }
        recalculateCombinedTransform()
    end
end)

room.on({"$ $ region $id at $x1 $y1 $x2 $y2 $x3 $y3 $x4 $y4",
         "$ $ region $id has name calendar"}, function(results)
    for i = 1, #results do
        local r = results[i]
        calendarRegion = {
            {x=r.x1, y=r.y1},
            {x=r.x2, y=r.y2},
            {x=r.x3, y=r.y3},
            {x=r.x4, y=r.y4}
        }
        recalculateCombinedTransform()
    end
end)

room.on({"$ $ draw graphics $graphics on web2"}, function(results)
    graphics_cache = {}
    for i = 1, #results do
        local parsedGraphics = json.decode(results[i].graphics)
        for g = 1, #parsedGraphics do
            graphics_cache[#graphics_cache + 1] = parsedGraphics[g]
        end
    end
end)

function love.load()
    love.window.setTitle('Room Graphics')
    love.window.setFullscreen( true )
    love.graphics.setBackgroundColor(0, 0, 0)
    font = love.graphics.newFont(72)
    room.init(true)
end

function love.draw()
    -- TODO: set baseline things
    is_fill_on = true
    fill_color = {255, 255, 255}
    is_stroke_on = true
    stroke_color = {255, 255, 255}
    stroke_width = 1
    love.graphics.setLineWidth( stroke_width )
    font_color = {255, 255, 255}
    local fontSize = 72
    love.graphics.setFont(font)

    love.graphics.replaceTransform(COMBINED_TRANSFORM)

    for i = 1, #graphics_cache do
        local g = graphics_cache[i]
        local opt = g.options
        if g.type == "rectangle" then
            if is_fill_on then
                love.graphics.setColor(fill_color)
                love.graphics.rectangle("fill", opt.x, opt.y, opt.w, opt.h)
            end
            if is_stroke_on then
                love.graphics.setColor(stroke_color)
                love.graphics.rectangle("line", opt.x, opt.y, opt.w, opt.h)
            end
        elseif g.type == "ellipse" then
            if is_fill_on then
                love.graphics.setColor(fill_color)
                love.graphics.ellipse("fill", opt.x + opt.w * 0.5, opt.y + opt.h * 0.5, opt.w * 0.5, opt.h * 0.5)
            end
            if is_stroke_on then
                love.graphics.setColor(stroke_color)
                love.graphics.ellipse("line", opt.x + opt.w * 0.5, opt.y + opt.h * 0.5, opt.w * 0.5, opt.h * 0.5)
            end
        elseif g.type == "line" then
            if is_stroke_on then
                love.graphics.setColor(stroke_color)
                love.graphics.line(opt[1], opt[2], opt[3], opt[4])
            end
        elseif g.type == "polygon" then
            local vertices = {}
            for j = 1, #opt do
                vertices[j*2 - 1] = opt[j][1]
                vertices[j*2] = opt[j][2]
            end
            if is_fill_on then
                love.graphics.setColor(fill_color)
                love.graphics.polygon('fill', vertices)
            end
            if is_stroke_on then
                love.graphics.setColor(stroke_color)
                love.graphics.polygon('line', vertices)
            end
        elseif g.type == "text" then
            love.graphics.setColor(font_color)
            local lineHeight = fontSize * 1.3
            for line in opt.text:gmatch("([^\n]*)\n?") do
                love.graphics.print(line, opt.x, opt.y + i * lineHeight)
            end
        elseif g.type == "fill" then
            is_fill_on = true
            if type(opt) == "string" then
                if colors[opt] ~= nill then
                    fill_color = colors[opt]
                end
            elseif #opt == 3 then
                fill_color = {opt[1], opt[2], opt[3]}
            elseif #opt == 4 then
                fill_color = {opt[1], opt[2], opt[3], opt[4]}
            end
        elseif g.type == "stroke" then
            is_stroke_on = true
            if type(opt) == "string" then
                if colors[opt] ~= nill then
                    stroke_color = colors[opt]
                end
            elseif #opt == 3 then
                stroke_color = {opt[1], opt[2], opt[3]}
            elseif #opt == 4 then
                stroke_color = {opt[1], opt[2], opt[3], opt[4]}
            end
        elseif g.type == "fontcolor" then
            if type(opt) == "string" then
                -- TODO: have color names like "red" in love2d
            elseif #opt == 3 then
                font_color = {opt[1], opt[2], opt[3]}
            elseif #opt == 4 then
                font_color = {opt[1], opt[2], opt[3], opt[4]}
            end
        elseif g.type == "nofill" then
            is_fill_on = false
        elseif g.type == "nostroke" then
            is_stroke_on = false
        elseif g.type == "strokewidth" then
            stroke_width = tonumber(opt)
            love.graphics.setLineWidth( stroke_width )
        elseif g.type == "fontsize" then
            font = love.graphics.newFont(tonumber(opt))
        elseif g.type == "push" then
            love.graphics.push()
        elseif g.type == "pop" then
            love.graphics.pop()
        elseif g.type == "translate" then
            love.graphics.translate(tonumber(opt.x), tonumber(opt.y))
        elseif g.type == "rotate" then
            love.graphics.rotate(tonumber(opt))
        elseif g.type == "scale" then
            love.graphics.scale(tonumber(opt.x), tonumber(opt.y))
        elseif g.type == "transform" then
            local elements = {}
            for j = 1, #opt do
                elements[j] = tonumber(opt[j])
            end
            -- interpret elements as row-major
            local transform = love.math.newTransform()
            transform.setMatrix("row", elements)
            love.graphics.applyTransform(transform)
        end
    end
end

function love.update()
    room.listen(true) -- blocking listen
end
