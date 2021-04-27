
(function() {
    //document.cookie = "lol=ok";

    let start_button = document.getElementById("pred_btn_start");
    let inputName = document.getElementById("pred_name");
    let inputOpt1 = document.getElementById("pred_opt1");
    let inputOpt2 = document.getElementById("pred_opt2");
    let inputDelay = document.getElementById("pred_delay");
    let predictionsListDiv = document.getElementById("betPredictionList");

    var createPredictionElement = function(message) {
        var div = document.createElement("div");
        div.appendChild(document.createTextNode("ok"));
        return div;
    }

    var ws;
    var open = function () {
        console.log("connecting to ws...")
        ws = new WebSocket("ws:" + location.host + "/echo/");
        ws.onclose = function (e) {
            console.log(e);
            setTimeout(open, 5000);
        }
        ws.onmessage = function (e) {
            var msg = JSON.parse(e.data);
            //TODO: handle parse error
            if (msg.subject === "NEW_PREDICTION") {
                predictionsListDiv.appendChild(createPredictionElement(msg));
            }
            console.log(msg);
        }
        ws.onopen = function (e) {
            console.log(e);
        }
    };
    open();



    let startPrediction = function(e) {
        ws.send(JSON.stringify({
            "subject": "CREATE_PREDICTION",
            "args": {
                name: inputName.value,
                opt1: inputOpt1.value,
                opt2: inputOpt2.value,
                delay: inputDelay.value,
            },
        }));
    }
    start_button.addEventListener("click", startPrediction);
}())