const N = 16;

for (var i = 0; i < 16; i += 1) {
    $("select").append(`<option value="${i}">${i}</option>`)
}

$('select').on('change', function () {
    $.ajax({
        type: "POST",
        url: "/cleanup-claim",
        data: {
            claim: `paper 1013 is pointing at paper ${this.value}`,
            retract: `$ paper 1013 is pointing at paper $`
        },
        success: function () { console.log("success") },
        failure: function (errMsg) { console.log(errMsg) }
    });
});