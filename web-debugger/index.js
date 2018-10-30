function b64DecodeUnicode(str) {
    // Going backwards: from bytestream, to percent-encoding, to original string.
    return decodeURIComponent(atob(str).split('').map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
}

$.get("http://localhost:3000/db", function (data) {
    console.log(data)
    var decodedData = data.map(b64DecodeUnicode)
    var dataJoinedBySource = {}
    decodedData.forEach(function (data) {
        var firstWord = data.split(" ")[0];
        if (!(firstWord in dataJoinedBySource)) {
            dataJoinedBySource[firstWord] = []    
        }
        dataJoinedBySource[firstWord].push(data)
    })
    console.log(decodedData)
    var decodedDataHTML = decodedData.map(function (data) {
        return `<li>${data}</li>`
    }).join('\n');

    var decodedDataHTML2 = ""
    Object.keys(dataJoinedBySource).forEach(function (source) {
        decodedDataHTML2 += `<h4>${source}</h4>`
        decodedDataHTML2 += dataJoinedBySource[source].map(function (data) {
            return `<li>${data}</li>`
        }).join('\n');
    })
    $(".results").html(decodedDataHTML2);
});