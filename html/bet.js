(function() {
    //document.cookie = "lol=ok";

    let ws;
    let currentUserId = -1;
    let start_button = document.getElementById("pred_btn_start");
    let inputName = document.getElementById("pred_name");
    let inputOpt1 = document.getElementById("pred_opt1");
    let inputOpt2 = document.getElementById("pred_opt2");
    let inputDelay = document.getElementById("pred_delay");
    let elPredictionsList = document.getElementById("betPredictionList");
    let elCreatePrediction = document.getElementById("betStartPrediction");
    let tickInterval = null;
    let gasAmountP = document.createElement("p");
    let gasAmountTextNode = document.createTextNode("0");
    let reButtonOdOpt = /bet_button_([0-9]+)_([12])/;
    let reButtonEndPrediction = /end_opt_([12])_([0-9]+)/;

    let predictions = {};

    gasAmountP.appendChild(gasAmountTextNode);

    let wsOnMessage = (e) => {
        let msg = JSON.parse(e.data);
        console.log(msg);
        //TODO: handle parse error
        if (msg.subject === "PREDICTION_STARTED") {
            predictions[msg.id] = new Prediction(msg);
        }
        if (msg.subject === "USER_INFO") {
            currentUserId = msg.args["Id"];
            if (msg.flags.includes("CAN_CREATE_PREDICTIONS")) {
                elCreatePrediction.style.display = "block";
            }
        }
        if (msg.subject === "PREDICTION_CHANGED") {
            predictions[msg.id].handleChange(msg);
        }
    }

    let Prediction = function (message) {
        this.id = message.args["id"];
        this.name = message.args["name"];
        this.infoRow = {};

        let countDown = getCountDown(message);
        this.table = document.createElement("table");
        this.table.setAttribute("cellspacing", "0");
        this.table.setAttribute("cellpadding", "0");
        let th = createElTextClass(this.table, "th", this.name, "");
        let cd = createElTextClass(th, "span", countDown, "predCountDown");
        cd.setAttribute("id", "predCountDown_" + this.id);
        th.setAttribute("colspan", 3);
        this.createBetInfoRow("G");
        this.createBetInfoRow("N");
        this.createBetInfoRow("P");
        this.createBetInfoRow("C");

        if (countDown < 1) {
            if (message.args.createdBy === currentUserId) {
                createEndPredictionRow(this.table, message.args.opt1, message.args.opt2, message.args.id);
            }
        } else {
            let c = document.createElement("span");
            let i1 = document.createElement("input");
            i1.setAttribute("id",`bet_amount_${message.args.id}_1`);
            i1.value = "10";
            i1.setAttribute("type", "number");

            let i2 = document.createElement("input");
            i2.setAttribute("id",`bet_amount_${message.args.id}_2`);
            i2.value = "10";
            i2.setAttribute("type", "number");
            createBetRow(this.table, i1, c, i2);
            createBetRow(this.table,
                createButton(message.args.opt1, `bet_button_${message.args.id}_1`, clickBetHandler),
                c,
                createButton(message.args.opt2, `bet_button_${message.args.id}_2`, clickBetHandler)
            );
        }
        elPredictionsList.appendChild(this.table);
        if (tickInterval === null) {
            tickInterval = setInterval(handleCountdownTick, 1000);
        }
    }

    Prediction.prototype.createBetInfoRow = function (name) {
        let tr = document.createElement("tr");
        this.infoRow[name] = [
            {
                textNode: document.createTextNode("0"),
                value: 0,
            }, {
                textNode: document.createTextNode("0"),
                value: 0,
            }
        ];
        createAppendEl(tr, "td", this.infoRow[name][0].textNode);
        createAppendEl(tr, "td", createElTextClass(tr, "span", name, ""));
        createAppendEl(tr, "td", this.infoRow[name][1].textNode);
        this.table.appendChild(tr);
    }

    Prediction.prototype.handleChange = function (msg) {

    }

    const createElTextClass = (parent, tagName, text, className) => {
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

    const clickEndPredictionHandler = (e) => {
        console.log(e);
        let idOpt = e.target.getAttribute("id").match(reButtonEndPrediction);
        let betId = idOpt[2];
        let opt1Won = idOpt[1] === "1";
        ws.send(JSON.stringify({
            subject: "PREDICTION_FINISHED",
            args: {
                id: betId,
                opt1Won: opt1Won ? "true" : "false",
            },
        }));
    }

    const clickBetHandler = (e) => {
        console.log(e);
        let idOpt = e.target.getAttribute("id").match(reButtonOdOpt);
        let betId = idOpt[1];
        let opt1Win = idOpt[2] === "1";
        let amountEl = document.getElementById(`bet_amount_${betId}_${idOpt[2]}`);
        console.log("sending");
        ws.send(JSON.stringify({
            subject: "BET",
            args: {
                id: betId,
                amount: amountEl.value,
                opt1Win: opt1Win ? "true" : "false",
            }
        }));
    }

    const handleCountdownTick = () => {
        let els = document.getElementsByClassName("predCountDown");
        if (els.length === 0) {
            clearInterval(tickInterval);
            tickInterval = null;
            return;
        }
        for (let i = 0; i < els.length; i++) {
            let count = parseInt(els[i].textContent);
            if (count === 1) {
                let strId = els[i].id.substring(14); // strlen "predCountDown_"
                document.getElementById(`bet_button_${strId}_1`).parentElement.parentElement.remove();
                document.getElementById(`bet_amount_${strId}_1`).parentElement.parentElement.remove();
            }
            if (count > 0) {
                els[i].textContent = "" + (count - 1);
            }
        }
    }

    const createAppendEl = (parent, tagName, el) => {
        let e = document.createElement(tagName);
        e.appendChild(el);
        parent.appendChild(e);
        return e;
    }

    const createBetRow = (table, el1, el2, el3) => {
        let tr = document.createElement("tr");
        createAppendEl(tr, "td", el1);
        createAppendEl(tr, "td", el2);
        createAppendEl(tr, "td", el3);
        table.appendChild(tr);
    }

    const createBetInfoRow = (table, name, id) => {

    }

    const getCountDown = (m) => {
        let delay = parseInt(m.args.delay, 10);
        let end = (Date.parse(m.args.createdAt) / 1000) + delay;
        let now = Date.now() / 1000;
        let cd = Math.round(end - now);
        if (cd < 0) {
            return 0;
        }
        return cd;
    }

    const createButton = (name, id, clickHandler) => {
        let b = document.createElement("button");
        b.appendChild(document.createTextNode(name));
        b.setAttribute("id", id);
        b.onclick = clickHandler;
        return b;
    }

    const createEndPredictionRow = (table, o1, o2, id) => {
        let c = document.createElement("span");
        let b1 = createButton(o1, "end_opt_1_" + id, clickEndPredictionHandler)
        let b2 = createButton(o2, "end_opt_2_" + id, clickEndPredictionHandler);
        createBetRow(table, b1, c, b2);
        return table;
    }

    const open = () => {
        console.log("connecting to ws...")

        ws = new WebSocket(`ws:${location.host}/echo/${window.location.search}`);
        ws.onclose = function (e) {
            console.log(e);
            setTimeout(open, 5000);
            elCreatePrediction.style.display = "none";
            while (elPredictionsList.firstChild) {
                elPredictionsList.removeChild(elPredictionsList.firstChild);
            }
        }
        ws.onmessage = wsOnMessage;
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
        if (delay > 900 || delay < 30) {
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
