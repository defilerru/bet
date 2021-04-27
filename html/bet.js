
(function() {
    //document.cookie = "lol=ok";

    let start_button = document.getElementById("pred_btn_start");
    let inputName = document.getElementById("pred_name");
    let inputOpt1 = document.getElementById("pred_opt1");
    let inputOpt2 = document.getElementById("pred_opt2");
    let inputDelay = document.getElementById("pred_delay");
    let predictionsListDiv = document.getElementById("betPredictionList");
    let tickInterval = null;

    let tickHandler = function() {
        for (let i = 0; i < predictionsListDiv.children.length; i++) {
            let pred = predictionsListDiv.children[i];
            for (let j = 0; j < pred.children.length; j++){
                c = pred.children[j];
                if (c.getAttribute("class") === "predCountDown") {
                    let count = parseInt(c.textContent);
                    if (count === 0) {
                        console.log("done");
                    } else {
                        c.textContent = count - 1;
                    }
                    console.log(c);
                }
            }
        }
    }

    let createPredictionElement = function(message) {
        let div = document.createElement("div");
        let nameP = document.createElement("p");
        nameP.setAttribute("class", "predName");
        let countDown = document.createElement("span");
        countDown.setAttribute("class", "predCountDown")
        countDown.appendChild(document.createTextNode(message.args.delay));
        nameP.appendChild(document.createTextNode(message.args.name));
        let option1P = document.createElement("button");
        option1P.appendChild(document.createTextNode(message.args.opt1));
        let option2P = document.createElement("button");
        option2P.appendChild(document.createTextNode(message.args.opt2));
        let amount = document.createElement("input");
        amount.setAttribute("type", "text");
        amount.setAttribute("value", "100");

        div.appendChild(countDown);
        div.appendChild(nameP);
        div.appendChild(amount);
        div.appendChild(option1P);
        div.appendChild(option2P);
        return div;
    }

    let ws;
    let open = function () {
        console.log("connecting to ws...")
        ws = new WebSocket("ws:" + location.host + "/echo/");
        ws.onclose = function (e) {
            console.log(e);
            setTimeout(open, 5000);
        }
        ws.onmessage = function (e) {
            var msg = JSON.parse(e.data);
            //TODO: handle parse error
            if (msg.subject === "PREDICTION_STARTED") {
                let de = predictionsListDiv.appendChild(createPredictionElement(msg));
                console.log(de);
            }
            console.log(msg);
            if (tickInterval === null) {
                tickInterval = setInterval(tickHandler, 1000);
                console.log(tickInterval);
            }
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