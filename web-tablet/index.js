function render(data) {
    if (!data || data.length === 0) {
        $(".results").html("No results")
        return
    }
    try {
        let text = JSON.parse(data)[0].text;
        $(".results").html(text);
    } catch {
        $(".results").html("error parsing");
    }
    
}

function update() {
    $.ajax({
        type: "GET",
        url: "/select",
        data: {
            subscription: JSON.stringify(["$ wish tablet would show $text"])
        },
        success: function (data) {
            console.log(data);
            render(data);
            setTimeout(update, 2000);
        },
        failure: function (errMsg) {
            render([]);
            alert(errMsg);
            setTimeout(update, 2000);
        }
    });
}

update();