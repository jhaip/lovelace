require "zhelpers"
local room = require "helper"

cwide = 520
chigh = 333
show_graphics = false

room.on({"$ $ I am a turtle card"}, function(results)
    if #results > 0 then
        show_graphics = true
    else
        show_graphics = false
    end
end)

function love.load()
    love.window.setFullscreen( true )
    love.window.setTitle(' Hello Wörld ')
    love.window.setMode(cwide, chigh)
    love.graphics.setBackgroundColor(0, 0, 0)
    font = love.graphics.setNewFont("FreeSans.ttf", 72)
    room.init(true)
end

function love.draw()
    if show_graphics then
        love.graphics.setColor(255, 255, 0)
        love.graphics.print("Hello World", cwide/4, chigh/3.33)
        love.graphics.circle("fill", cwide/2, chigh/2, 50)
    end
end

function love.update()
    room.listen(true) -- blocking listen
end
