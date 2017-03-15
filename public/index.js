const tickerMsg = JSON.stringify({
    event: 'subscribe',
    channel: 'ticker',
    symbol: 'tBTCUSD'
});

const tradesMsg = JSON.stringify({
    event: 'subscribe',
    channel: 'trades',
    symbol: 'tBTCUSD'
});

if (window.attachEvent) {
    window.attachEvent('onload', () => {
        loadSocket(tickerMsg);
        loadSocket(tradesMsg);
    }, false);
} else if (window.addEventListener) {
    window.addEventListener('load', () => {
        loadSocket(tickerMsg);
        loadSocket(tradesMsg);
    }, false);
} else {
    document.addEventListener('load', () => {
        loadSocket(tickerMsg);
        loadSocket(tradesMsg);
    }, false);
}



function loadSocket(msg) {
    const ws = new WebSocket('wss://api.bitfinex.com/ws/2');
    let chanID;
    let obj;

    ws.onmessage = res => {
        const data = JSON.parse(res.data);
        console.log(data);
        if (['hb', 'te'].indexOf(data[1]) <= -1) {
            chanID = data[0];
            console.log(data[1]);
            const s = JSON.parse(msg);
            switch (s.channel) {
                case 'ticker':
                    obj = parseWebsocketToObject(data, chanID)
                    updateTickDataOnPage(obj);
                    break;
                case 'trades':
                    obj = parseTradesToObject(data[2], chanID);
                    console.log(obj);
                    displayTradeDataToPage(obj);
                    break;

            }

            //buildLineChart(data.curTime, data.bid)
        }
    }

    ws.onerror = () => {
        const unsubscribe = JSON.stringify({
            event: 'subscribe',
            chanId: chanID
        });
        ws.send(unsubscribe);
        ws.close();

    }

    ws.onopen = () => {
        ws.send(msg);
    }
}

function parseWebsocketToObject(data, chanID) {
    const curTime = getCurrentDate();
    const jsonData = { bid: data[1][0], time: curTime, chanID, ask: data[1][3], price: data[1][2], low: data[1][9], high: data[1][8], change: data[1][5] };
    return jsonData;
}

function parseTradesToObject(data, chanID) {
    const curTime = getCurrentDate();
    const jsonData = { MTS: data[1], amount: data[2], price: data[3], chanID, time: curTime };
    console.log(jsonData);
    return jsonData;
}

function displayTradeDataToPage(jsonData) {
    const trade = document.createElement('div');
    trade.className = 'depthcell';
    trade.innerHTML = `<span class="depthType" id="depthAsk1">Ask:</span><span class="depthPrice" >${jsonData.price}</span><span class="depthVol">${jsonData.amount}</span>`;
    document.getElementById('orders2').appendChild(trade);
}

function getCurrentDate() {
    const today = new Date();
    const day = today.getDay();
    const daylist = ["Sunday", "Monday", "Tuesday", "Wednesday ", "Thursday", "Friday", "Saturday"];
    let hour = today.getHours();
    const minute = today.getMinutes();
    const second = today.getSeconds();
    let prepand = (hour >= 12) ? " PM " : " AM ";
    hour = (hour >= 12) ? hour - 12 : hour;
    if (hour === 0 && prepand === ' PM ') {
        if (minute === 0 && second === 0) {
            hour = 12;
            prepand = ' Noon';
        } else {
            hour = 12;
            prepand = ' PM';
        }
    }
    if (hour === 0 && prepand === ' AM ') {
        if (minute === 0 && second === 0) {
            hour = 12;
            prepand = ' Midnight';
        } else {
            hour = 12;
            prepand = ' AM';
        }
    }
    const time = `${hour} : ${minute} : ${second} ${prepand}`;
    return time;
}

function buildLineChart(xdata, ydata) {
    const chart = new CanvasJS.Chart("chartContainer", {

        title: {
            text: "Bitfenix Tick Data"
        },
        data: [{
            type: "line",

            dataPoints: [
                { x: xdata, y: ydata }
            ]
        }]
    });

    chart.render();
}

function updateTickDataOnPage(obj) {
    document.getElementById('buytext').innerHTML = `Bid: ${obj.bid}`;
    document.getElementById('sell').innerHTML = obj.ask;
    document.getElementById('lastTrade').innerHTML = obj.price;
    document.getElementById('low').innerHTML = obj.low;
    document.getElementById('high').innerHTML = obj.high;
    document.getElementById('centval').innerHTML = obj.change;
}