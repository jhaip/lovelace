require "zhelpers"
local room = require "helper"
local json = require "json"

cwide = 520
chigh = 333
show_graphics = false
graphics_cache = {}
font = false

room.on({"$ $ I am a turtle card"}, function(results)
    if #results > 0 then
        show_graphics = true
    else
        show_graphics = false
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
    love.window.setTitle(' Hello Wörld ')
    love.window.setMode(cwide, chigh)
    love.window.setFullscreen( true )
    love.graphics.setBackgroundColor(0, 0, 0)
    font = love.graphics.newFont(72)
    room.init(true)
end

function love.draw()
    if show_graphics then
        love.graphics.setColor(255, 255, 0)
        love.graphics.print("Hello World", cwide/4, chigh/3.33)
        love.graphics.circle("fill", cwide/2, chigh/2, 50)
    end
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
                love.graphics.ellipse("fill", opt.x, opt.y, opt.w * 0.5, opt.h * 0.5)
            end
            if is_stroke_on then
                love.graphics.setColor(stroke_color)
                love.graphics.ellipse("line", opt.x, opt.y, opt.w * 0.5, opt.h * 0.5)
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
                -- TODO: have color names like "red" in love2d
            elseif #opt == 3 then
                fill_color = {opt[1], opt[2], opt[3]}
            elseif #opt == 4 then
                fill_color = {opt[1], opt[2], opt[3], opt[4]}
            end
        elseif g.type == "stroke" then
            is_stroke_on = true
            if type(opt) == "string" then
                -- TODO: have color names like "red" in love2d
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
            local transform = Transform:setMatrix("row", elements)
            love.graphics.replaceTransform(transform)
        end
    end
end

function love.update()
    room.listen(true) -- blocking listen
end
