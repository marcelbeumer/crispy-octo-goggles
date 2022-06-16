let serverData = {
  message: "ok",
  points: [
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 0, type: "low" },
    { time: 1655403669818, count: 65, type: "low" },
    { time: 1655403669818, count: 159, type: "low" },
    { time: 1655403669818, count: 29, type: "low" },
    { time: 1655406600000, count: 70, type: "high" },
    { time: 1655406900000, count: 146, type: "high" },
    { time: 1655407200000, count: 37, type: "high" },
  ],
};

let dataHigh = [];
let dataLow = [];
for (const p of serverData.points) {
  var o = { x: p.time, y: p.count };
  if (p.type === "low") {
    dataLow.push(o);
  } else {
    dataHigh.push(o);
  }
}

let colors = ["#ED6D85", "#57A0E5", "#6DBDBF"];

let el = document.getElementById("chart");
let ctx = el.getContext("2d");
let myChart = new Chart(ctx, {
  type: "line",
  data: {
    datasets: [
      {
        label: "Low",
        backgroundColor: colors[0],
        borderColor: colors[0],
        cubicInterpolationMode: "monotone",
        tension: 0.4,
        data: dataLow,
      },
      {
        label: "High",
        backgroundColor: colors[1],
        borderColor: colors[1],
        cubicInterpolationMode: "monotone",
        tension: 0.4,
        data: dataHigh,
      },
    ],
  },
  options: {
    scales: {
      x: {
        type: "time",
        min: 1655403669818 - 1000 * 60,
        max: 1655407200000 + 1000 * 60,
      },
    },
  },
});
