/**
 * Original code copied from https://github.com/spotify/web-api-auth-examples/blob/master/authorization_code/app.js
 */

// TODO notes from 2020-08-27 night:
// * claim general status to the room
// * save access token and refresh token in global context so it can be used everywhere
// * save state of access token / refresh token
// * add a method to get a new access token using refresh token and refresh API method
//   and otherwise ask user to log in via web site

const { room, myId, run } = require('../helper2')(__filename);
var express = require('express'); // Express web server framework
var request = require('request'); // "Request" library
var cors = require('cors');
var querystring = require('querystring');
var cookieParser = require('cookie-parser');
const { ZMQ_ROUTER_MANDATORY } = require('zeromq');

var client_id = process.env.SPOTIFY_CLIENT_ID; // Your client id
var client_secret = process.env.SPOTIFY_CLIENT_SECRET; // Your secret
var redirect_uri = 'localhost:8888/callback'; // Your redirect uri

/**
 * Generates a random string containing numbers and letters
 * @param  {number} length The length of the string
 * @return {string} The generated string
 */
var generateRandomString = function (length) {
    var text = '';
    var possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';

    for (var i = 0; i < length; i++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length));
    }
    return text;
};

var stateKey = 'spotify_auth_state';

var app = express();

app.use(express.static(__dirname + '/public'))
    .use(cors())
    .use(cookieParser());

app.get('/login', function (req, res) {

    var state = generateRandomString(16);
    res.cookie(stateKey, state);

    // your application requests authorization
    var scope = 'user-read-private user-read-email';
    res.redirect('https://accounts.spotify.com/authorize?' +
        querystring.stringify({
            response_type: 'code',
            client_id: client_id,
            scope: scope,
            redirect_uri: redirect_uri,
            state: state
        }));
});

app.get('/callback', function (req, res) {

    // your application requests refresh and access tokens
    // after checking the state parameter

    var code = req.query.code || null;
    var state = req.query.state || null;
    var storedState = req.cookies ? req.cookies[stateKey] : null;

    if (state === null || state !== storedState) {
        res.redirect('/#' +
            querystring.stringify({
                error: 'state_mismatch'
            }));
    } else {
        res.clearCookie(stateKey);
        var authOptions = {
            url: 'https://accounts.spotify.com/api/token',
            form: {
                code: code,
                redirect_uri: redirect_uri,
                grant_type: 'authorization_code'
            },
            headers: {
                'Authorization': 'Basic ' + (new Buffer(client_id + ':' + client_secret).toString('base64'))
            },
            json: true
        };

        request.post(authOptions, function (error, response, body) {
            if (!error && response.statusCode === 200) {

                var access_token = body.access_token,
                    refresh_token = body.refresh_token;

                var options = {
                    url: 'https://api.spotify.com/v1/me',
                    headers: { 'Authorization': 'Bearer ' + access_token },
                    json: true
                };

                // use the access token to access the Spotify Web API
                request.get(options, function (error, response, body) {
                    console.log(body);
                });

                // we can also pass the token to the browser to make requests from there
                res.redirect('/#' +
                    querystring.stringify({
                        access_token: access_token,
                        refresh_token: refresh_token
                    }));
            } else {
                res.redirect('/#' +
                    querystring.stringify({
                        error: 'invalid_token'
                    }));
            }
        });
    }
});

app.get('/refresh_token', function (req, res) {

    // requesting access token from refresh token
    var refresh_token = req.query.refresh_token;
    var authOptions = {
        url: 'https://accounts.spotify.com/api/token',
        headers: { 'Authorization': 'Basic ' + (new Buffer(client_id + ':' + client_secret).toString('base64')) },
        form: {
            grant_type: 'refresh_token',
            refresh_token: refresh_token
        },
        json: true
    };

    request.post(authOptions, function (error, response, body) {
        if (!error && response.statusCode === 200) {
            var access_token = body.access_token;
            res.send({
                'access_token': access_token
            });
        }
    });
});

function updateCurrentlyPlaying() {
    var options = {
        url: 'https://api.spotify.com/v1/me/player/currently-playing',
        headers: { 'Authorization': 'Bearer ' + access_token },
        json: true
    };

    // use the access token to access the Spotify Web API
    request.get(options, function (error, response, body) {
        if (!!error || response.statusCode !== 200) {
            console.log("ERROR")
            console.log(error);
            console.log(response)
            // TODO: log error to room
            if (response.statusCode === 401) {
                // TODO:
                // Fetch new access token using refresh token
                // retry updateCurrentlyPlaying() again
            }
        } else {
            if (response.statusCode === 204) {
                room.assert(`currently playing Spotify song is nothing @ ${currentTimeMs}`);
            } else {
                // console.log(body);
                const songTitle = body.item.name;
                const songArtist = body.item.artists.map(a => a.name).join(", ");
                const currentTimeMs = (new Date()).getTime()
                room.assert(`currently playing Spotify song is`, ["text", songTitle], `by`, ["text", songArtist], `@ ${currentTimeMs}`);
            }
        }
    });
}

function playSpotifyUri(uri) {
    var jsonBody = {};
    if (uri.includes("track")) {
        jsonBody = {
            "uris": [uri] // ex: "spotify:track:232Y5lAjifWnc5NhUn34mz"
        }
    } else {
        jsonBody = { "context_uri": uri } // ex: "spotify:album:04HMMwLmjkftjWy7xc6Bho"
    }
    var options = {
        url: 'https://api.spotify.com/v1/me/player/play',
        headers: { 'Authorization': 'Bearer ' + access_token },
        method: 'PUT',
        json: jsonBody
    };

    // use the access token to access the Spotify Web API
    request(options, function (error, response, body) {
        if (!!error || response.statusCode !== 200) {
            console.log("ERROR")
            console.log(error);
            console.log(response)
            // TODO: log error to room
            if (response.statusCode === 401) {
                // TODO:
                // Fetch new access token using refresh token
                // retry updateCurrentlyPlaying() again
            }
            if (response.statusCode === 404) {
                console.log("Device probably needs to be selected")
                // TOOD: log this to the room
            }
        } else {
            // TODO: clear wish to play song?
            console.log("successfully started playing song")
        }
    });
}

room.on(
    `wish currently Spotify song would be updated`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results) {
            updateCurrentlyPlaying();
        }
        room.subscriptionPostfix();
    });

room.on(
    `wish $spotifyUri would be played on Spotify`,
    results => {
        room.subscriptionPrefix(2);
        if (!!results && results.length > 0) {
            playSpotifyUri(results[0].spotifyUri);
        }
        room.subscriptionPostfix();
    });

console.log('Listening on 8888');
app.listen(8888);