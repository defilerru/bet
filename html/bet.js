
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
    let gasAmountP = document.createElement("p");
    let gasAmountTextNode = document.createTextNode("0");
    gasAmountP.appendChild(gasAmountTextNode);

    let createElTextClass = function(parent, tagName, text, className) {
        let el = document.createElement(tagName);
        el.appendChild(document.createTextNode(text));
        if (className !== "") {
            el.setAttribute("class", className);
        }
        return parent.appendChild(el);
    }

    let createAndAttach = function(parent, tagName, attributes) {
        let el = document.createElement(tagName);
        parent.appendChild(el);
        for (const [attr, value] of Object.entries(attributes)) {
            el.setAttribute(attr, value);
        }
        return el
    }

    let getChildByClass = function(element, className) {
        for (let i = 0; i < element.children.length; i++) {
            let el = element.children[i];
            if (el.getAttribute("class") === className) {
                return el;
            }
        }
    }

    const bet = (e) => {
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

    const tickHandler = () => {
        for (let i = 0; i < predictionsListDiv.children.length; i++) {
            let pred = predictionsListDiv.children[i];
            let c = getChildByClass(pred, "predCountDown")
            let count = parseInt(c.textContent);
            if (count === 0) {
                //console.log("done");
            } else {
                c.textContent = "" + (count - 1);
            }
        }
    }

    const createAppendEl = (parent, tagName, el) => {
        let e = document.createElement(tagName);
        e.appendChild(el);
        parent.appendChild(e);
        return e;
    }

    const createBetRow = (table, name, el1, el2) => {
        let tr = document.createElement("tr");
        let td1 = createAppendEl(tr, "td", el1);
        let tdc = createAppendEl(tr, "td", createElTextClass(tr, "span", name, ""));
        let td2 = createAppendEl(tr, "td", el2);
        table.appendChild(tr);
    }

    const createBetInfoRow = (table, name) => {
        let tr = document.createElement("tr");
        let td1 = createAppendEl(tr, "td", createElTextClass(tr, "span", "-", name + "_1"));
        let tdc = createAppendEl(tr, "td", createElTextClass(tr, "span", name, ""));
        let td2 = createAppendEl(tr, "td", createElTextClass(tr, "span", "-", name + "_2"));
        table.appendChild(tr);
    }

    const createPredictionElement = (message) => {
        let table = document.createElement("table");
        table.setAttribute("cellspacing", "0");
        table.setAttribute("cellpadding", "0");
        let th = createElTextClass(table, "th", message.args.name, "");
        th.setAttribute("colspan", 3);
        createBetInfoRow(table, "G");
        createBetInfoRow(table, "#");
        createBetInfoRow(table, "%");
        createBetInfoRow(table, "/");
        return table;
    }

    const open = () => {
        console.log("connecting to ws...")

        ws = new WebSocket("ws:" + location.host + "/echo/"+window.location.search);
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
                //tickInterval = setInterval(tickHandler, 1000);
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

    const startPrediction = (e) => {
        if (inputName.value === "" || inputOpt1.value === "" || inputOpt2.value === "") {
            //TODO: display error
            console.log("start prediction: invalid input");
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
