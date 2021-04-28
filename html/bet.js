
(function() {
    //document.cookie = "lol=ok";

    let ws;
    let start_button = document.getElementById("pred_btn_start");
    let inputName = document.getElementById("pred_name");
    let inputOpt1 = document.getElementById("pred_opt1");
    let inputOpt2 = document.getElementById("pred_opt2");
    let inputDelay = document.getElementById("pred_delay");
    let predictionsListDiv = document.getElementById("betPredictionList");
    let tickInterval = null;

    let createElTextClass = function(parent, tagName, text, className) {
        let el = document.createElement(tagName);
        el.appendChild(document.createTextNode(text));
        if (className !== "") {
            el.setAttribute("class", className);
        }
        return parent.appendChild(el);
    }

    let getChildByClass = function(element, className) {
        for (let i = 0; i < element.children.length; i++) {
            let el = element.children[i];
            if (el.getAttribute("class") === className) {
                return el;
            }
        }
    }

    let bet = function (e) {
        let idEl = getChildByClass(e.target.parentElement, "predictionId");
        let amountEl = getChildByClass(e.target.parentElement, "predAmount");
        ws.send(JSON.stringify({
            subject: "BET",
            args: {
                id: idEl.textContent,
                amount: amountEl.value,
                opt1Win: "" + (e.target.className === "betOpt1")}
        }));
        console.log(e.target);
    }

    let tickHandler = function() {
        for (let i = 0; i < predictionsListDiv.children.length; i++) {
            let pred = predictionsListDiv.children[i];
            let c = getChildByClass(pred, "predCountDown")
            let count = parseInt(c.textContent);
            if (count === 0) {
                //console.log("done");
            } else {
                c.textContent = count - 1;
            }
        }
    }

    let createPredictionElement = function(message) {
        let div = document.createElement("div");
        let pid = createElTextClass(div, "span", message.args.id, "predictionId");
        pid.setAttribute("hidden", "true"); //why css doesn't work?
        createElTextClass(div, "span", message.args.delay, "predCountDown");
        createElTextClass(div, "p", message.args.name, "predName");

        let option1B = createElTextClass(div, "button", message.args.opt1, "betOpt1");
        let option2B = createElTextClass(div, "button", message.args.opt2, "betOpt2");
        option1B.addEventListener("click", bet);
        option2B.addEventListener("click", bet);

        let amount = document.createElement("input");
        amount.setAttribute("class", "predAmount");
        amount.setAttribute("type", "text");
        amount.setAttribute("value", "100");
        div.appendChild(amount);

        return div;
    }

    let open = function () {
        console.log("connecting to ws...")
        ws = new WebSocket("ws:" + location.host + "/echo/");
        ws.onclose = function (e) {
            console.log(e);
            setTimeout(open, 5000);
        }
        ws.onmessage = function (e) {
            let msg = JSON.parse(e.data);
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
        ws.onerror = function (e) {
            console.log(e);
        }
    };
    open();

    let startPrediction = function(e) {
        if (inputName.value === "" || inputOpt1.value === "" || inputOpt2.value === "") {
            return;
        }
        let delay = parseInt(inputDelay.value);
        if (delay > 300 || delay < 30) {
            return;
        }
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