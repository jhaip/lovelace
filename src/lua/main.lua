-- cp /usr/share/fonts/truetype/freefont/FreeSans.ttf .
-- zip hello.love main.lua FreeSans.ttf
-- love ./hello.love

cwide = 520
chigh = 333

love.window.setTitle(' Hello Wörld ')
love.window.setMode(cwide, chigh)

function love.load()
        love.graphics.setBackgroundColor(177, 106, 248)
        font = love.graphics.setNewFont("FreeSans.ttf", 72)
end

function love.draw()
love.graphics.setColor(0, 0, 0)
        love.graphics.print("Hello World", cwide/4, chigh/3.33)
        love.graphics.circle("fill", cwide/2, chigh/2, 50)
end