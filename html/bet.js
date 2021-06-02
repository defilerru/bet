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
    let gasAmountP = document.createElement("p");
    let gasAmountTextNode = document.createTextNode("0");
    let reButtonEndPrediction = /end_opt_([12])_([0-9]+)/;

    let predictions = {};

    gasAmountP.appendChild(gasAmountTextNode);

    let wsOnMessage = (e) => {
        let msg = JSON.parse(e.data);
        console.log(msg);
        //TODO: handle parse error
        if (msg.subject === "PREDICTION_STARTED") {
            predictions[msg.args.id] = new Prediction(msg);
        }
        if (msg.subject === "PREDICTION_FINISHED") {
            predictions[msg.args.id].handleFinished(msg);
        }
        if (msg.subject === "USER_INFO") {
            currentUserId = msg.args.id;
            if (msg.flags.includes("CAN_CREATE_PREDICTIONS")) {
                elCreatePrediction.style.display = "block";
            }
        }
        if (msg.subject === "PREDICTION_CHANGED") {
            predictions[msg.args.id].handleChange(msg);
        }
    }

    let Prediction = function (message) {
        this.args = message.args;
        this.id = message.args["id"];
        this.infoRow = {};
        this.countDown = getCountDown(message);
        this.countDownTextNode = document.createTextNode("" + this.countDown);
        this.table = document.createElement("table");
        this.table.setAttribute("cellspacing", "0");
        this.table.setAttribute("cellpadding", "0");
        let nameP = document.createElement("p");
        nameP.appendChild(document.createTextNode(message.args["name"]));
        this.createFullRow(nameP, "predictionName");
        let countDownSpan = document.createElement("span");
        countDownSpan.appendChild(this.countDownTextNode);
        this.countDownElement = this.createFullRow(countDownSpan, "predictionCountdown");

        this.opt1El = createElTextClass(this.table, "span", this.args["opt1"]);
        this.opt2El = createElTextClass(this.table, "span", this.args["opt2"]);
        this.createBetRow(this.opt1El, null, this.opt2El);

        this.createBetInfoRow("G");
        this.createBetInfoRow("N");
        this.createBetInfoRow("P");
        this.createBetInfoRow("C");

        if (this.countDown < 1) {
            if (message.args["createdBy"] === currentUserId) {
                this.createEndPredictionRow();
            }
        } else {
            let i1 = document.createElement("input");
            i1.setAttribute("id",`bet_amount_${message.args.id}_1`);
            i1.value = "10";
            i1.setAttribute("type", "number");

            let i2 = document.createElement("input");
            i2.setAttribute("id",`bet_amount_${message.args.id}_2`);
            i2.value = "10";
            i2.setAttribute("type", "number");
            this.betRowInputEl = this.createBetRow(i1, null, i2);
            let b1 = createButton(message.args.opt1, "", (e) => {this.handleBet("true", i1)});
            let b2 = createButton(message.args.opt2, "", (e) => {this.handleBet("false", i2)});
            this.betRowButtonEl = this.createBetRow(b1, null,b2,);
            this.tickInterval = setInterval(() => {this.handleCountdownTick();}, 1000);
        }
        elPredictionsList.appendChild(this.table);
    }

    Prediction.prototype.createFullRow = function (el, css) {
        let td = document.createElement("td");
        td.setAttribute("colspan", "3");
        td.appendChild(el);
        let tr = document.createElement("tr");
        tr.classList.add(css);
        tr.appendChild(td);
        this.table.appendChild(tr);
        return tr;
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

    Prediction.prototype.handleChange = function(msg) {
        //TODO
    }

    Prediction.prototype.handleFinished = function(msg) {
        let el = (msg.args["opt1Won"] === "true") ? this.opt1El : this.opt2El;
        el.classList.add("predictionWinner");
        let td = document.createElement("td");
        td.setAttribute("colspan", "3");
        td.appendChild(createButton("dismiss", "", () => {this.handleClose()}));
        createAppendEl(this.table, "tr", td);
    }

    Prediction.prototype.handleClose = function() {
        this.table.remove();
        delete predictions[this.id];
    }

    Prediction.prototype.handleBet = function(opt1Win, amountEl) {
        ws.send(JSON.stringify({
            subject: "BET",
            args: {
                id: this.id,
                amount: amountEl.value,
                opt1Win: opt1Win,
            }
        }));
        this.betRowInputEl.remove()
        this.betRowButtonEl.remove()
    }

    Prediction.prototype.handleCountdownTick = function() {
        this.countDown--;
        this.countDownTextNode.nodeValue = "" + this.countDown;
        if (this.countDown < 0) {
            this.betRowInputEl.remove();
            this.betRowButtonEl.remove();
            this.countDownElement.remove();
            clearInterval(this.tickInterval);
            if (this.args["createdBy"] === currentUserId) {
                this.createEndPredictionRow();
            }
        }
    }

    Prediction.prototype.createEndPredictionRow = function() {
        let b1 = createButton(this.args.opt1, "", () => {this.handleEndPredictionClick("true")});
        let b2 = createButton(this.args.opt2, "", () => {this.handleEndPredictionClick("false")});
        this.createBetRow(b1, null, b2);
        return this.table;
    }

    Prediction.prototype.handleEndPredictionClick = function (opt1Won) {
        ws.send(JSON.stringify({
            subject: "PREDICTION_FINISHED",
            args: {
                id: this.id,
                opt1Won: opt1Won,
            },
        }));
    }

    Prediction.prototype.createBetRow = function (el1, elc, el2) {
        let tr = document.createElement("tr");
        createAppendEl(tr, "td", el1);
        createAppendEl(tr, "td", elc);
        createAppendEl(tr, "td", el2);
        this.table.appendChild(tr);
        return tr;
    }

    const createElTextClass = (parent, tagName, text, className) => {
        let el = document.createElement(tagName);
        el.appendChild(document.createTextNode(text));
        if (className !== "") {
            el.setAttribute("class", className);
        }
        return parent.appendChild(el);
    }

    const createAppendEl = (parent, tagName, el) => {
        let e = document.createElement(tagName);
        if (el !== null) {
            e.appendChild(el);
        }
        parent.appendChild(e);
        return e;
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
        if (id !== "") {
            b.setAttribute("id", id);
        }
        b.onclick = clickHandler;
        return b;
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
        if (delay > 900 || delay < 3) {
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
