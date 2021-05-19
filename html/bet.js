
(function() {
    //document.cookie = "lol=ok";

    let ws;
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
    let buttonOdOptRe = /bet_button_([0-9]+)_([12])/;
    gasAmountP.appendChild(gasAmountTextNode);

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

    const setTextPredInfo = (id, val) => {
        document.getElementById(id).firstChild.textContent = val;
    }

    const handlePredictionChanged = (msg) => {
        setTextPredInfo("betInfo1_G_" + msg.args.id, msg.args.amountOpt1);
        setTextPredInfo("betInfo2_G_" + msg.args.id, msg.args.amountOpt2);

        setTextPredInfo("betInfo1_N_" + msg.args.id, msg.args.ppl1);
        setTextPredInfo("betInfo2_N_" + msg.args.id, msg.args.ppl2);

        setTextPredInfo("betInfo1_C_" + msg.args.id, msg.args.coef1);
        setTextPredInfo("betInfo2_C_" + msg.args.id, msg.args.coef2);

        setTextPredInfo("betInfo1_P_" + msg.args.id, msg.args.per1);
        setTextPredInfo("betInfo2_P_" + msg.args.id, msg.args.per2);
    }

    const betClickHandler = (e) => {
        console.log(e);
        let idOpt = e.target.getAttribute("id").match(buttonOdOptRe);
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
            let count = parseInt(els[i].textContent) - 1;
            if (count <= 0) {
                let strId = els[i].id.substring(14); // strlen "predCountDown_"
                document.getElementById(`bet_button_${strId}_1`).parentElement.parentElement.remove();
                //TODO: remove input also
            } else {
                els[i].textContent = "" + count;
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
        let tr = document.createElement("tr");
        let s1 = createElTextClass(tr, "span", "-", name + "_1");
        let s2 = createElTextClass(tr, "span", "-", name + "_2");
        s1.setAttribute("id",`betInfo1_${name}_${id}`);
        s2.setAttribute("id",`betInfo2_${name}_${id}`);
        createAppendEl(tr, "td", s1);
        createAppendEl(tr, "td", createElTextClass(tr, "span", name, ""));
        createAppendEl(tr, "td", s2);
        table.appendChild(tr);
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

    const createPredictionElement = (message) => {
        let countDown = getCountDown(message);
        let table = document.createElement("table");
        table.setAttribute("cellspacing", "0");
        table.setAttribute("cellpadding", "0");
        let th = createElTextClass(table, "th", message.args.name, "");
        let cd = createElTextClass(th, "span", countDown, "predCountDown");
        cd.setAttribute("id", "predCountDown_" + message.args.id);
        th.setAttribute("colspan", 3);
        createBetInfoRow(table, "G", message.args.id);
        createBetInfoRow(table, "N", message.args.id);
        createBetInfoRow(table, "P", message.args.id);
        createBetInfoRow(table, "C", message.args.id);

        let c = document.createElement("span");

        if (countDown <= 0) {
            return table;
        }

        let i1 = document.createElement("input");
        i1.setAttribute("id",`bet_amount_${message.args.id}_1`);
        i1.value = "10";
        i1.setAttribute("type", "number");

        let i2 = document.createElement("input");
        i2.setAttribute("id",`bet_amount_${message.args.id}_2`);
        i2.value = "10";
        i2.setAttribute("type", "number");
        createBetRow(table, i1, c, i2);

        let b1 = document.createElement("button");
        b1.appendChild(document.createTextNode(message.args.opt1));
        b1.setAttribute("id", `bet_button_${message.args.id}_1`);

        let b2 = document.createElement("button");
        b2.appendChild(document.createTextNode(message.args.opt2));
        b2.setAttribute("id", `bet_button_${message.args.id}_2`);
        createBetRow(table, b1, c, b2);

        b1.onclick = betClickHandler;
        b2.onclick = betClickHandler;

        return table;
    }

    const open = () => {
        console.log("connecting to ws...")

        ws = new WebSocket("ws:" + location.host + "/echo/"+window.location.search);
        ws.onclose = function (e) {
            console.log(e);
            setTimeout(open, 5000);
            elCreatePrediction.style.display = "none";
            while (elPredictionsList.firstChild) {
                elPredictionsList.removeChild(elPredictionsList.firstChild);
            }
        }
        ws.onmessage = function (e) {
            let msg = JSON.parse(e.data);
            console.log(msg);
            //TODO: handle parse error
            if (msg.subject === "PREDICTION_STARTED") {
                let de = elPredictionsList.appendChild(createPredictionElement(msg));
                if (tickInterval === null) {
                    tickInterval = setInterval(handleCountdownTick, 1000);
                }
            }
            if (msg.subject === "USER_INFO") {
                if (msg.flags.includes("CAN_CREATE_PREDICTIONS")) {
                    elCreatePrediction.style.display = "block";
                }
            }
            if (msg.subject === "PREDICTION_CHANGED") {
                handlePredictionChanged(msg);
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
