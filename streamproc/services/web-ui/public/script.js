function getScales() {
  const padding = 0;
  const end = Date.now();
  const start = end - 1000 * 60 * 10;
  const min = start - padding;
  const max = end + padding;
  return { x: { type: "time", min, max } };
}

function createChart() {
  let colors = ["#ED6D85", "#57A0E5", "#6DBDBF"];
  let el = document.getElementById("chart");
  let ctx = el.getContext("2d");
  return new Chart(ctx, {
    type: "line",
    data: {
      datasets: [
        {
          label: "Low",
          backgroundColor: colors[0],
          borderColor: colors[0],
          cubicInterpolationMode: "monotone",
          tension: 0.4,
          data: [],
        },
        {
          label: "High",
          backgroundColor: colors[1],
          borderColor: colors[1],
          cubicInterpolationMode: "monotone",
          tension: 0.4,
          data: [],
        },
      ],
    },
    options: {
      scales: getScales(),
    },
  });
}

async function updateData(chart) {
  const res = await fetch("/api/data");
  const json = await res.json();
  if (json.message !== "ok") {
    throw new Error("fetched server data not ok");
  }
  const dataHigh = [];
  const dataLow = [];
  for (const p of json.points) {
    var o = { x: p.time, y: p.count };
    if (p.type === "low") {
      dataLow.push(o);
    } else {
      dataHigh.push(o);
    }
  }
  chart.data.datasets[0].data = dataLow;
  chart.data.datasets[1].data = dataHigh;
  chart.update();
}

async function main() {
  const chart = createChart();
  await updateData(chart).catch(console.warn);
  setInterval(() => {
    updateData(chart).catch(console.warn);
  }, 5000);
}

main();

//
// function addData(chart, label, data) {
//     chart.data.labels.push(label);
//     chart.data.datasets.forEach((dataset) => {
//         dataset.data.push(data);
//     });
//     chart.update();
// }
//
// function removeData(chart) {
//     chart.data.labels.pop();
//     chart.data.datasets.forEach((dataset) => {
//         dataset.data.pop();
//     });
//     chart.update();
// }

// let serverData = {
//   message: "ok",
//   points: [
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 0, type: "low" },
//     { time: 1655403669818, count: 65, type: "low" },
//     { time: 1655403669818, count: 159, type: "low" },
//     { time: 1655403669818, count: 29, type: "low" },
//     { time: 1655406600000, count: 70, type: "high" },
//     { time: 1655406900000, count: 146, type: "high" },
//     { time: 1655407200000, count: 37, type: "high" },
//   ],
// };
//
// let dataHigh = [];
// let dataLow = [];
// for (const p of serverData.points) {
//   var o = { x: p.time, y: p.count };
//   if (p.type === "low") {
//     dataLow.push(o);
//   } else {
//     dataHigh.push(o);
//   }
// }
//
// let colors = ["#ED6D85", "#57A0E5", "#6DBDBF"];
// let el = document.getElementById("chart");
// let ctx = el.getContext("2d");
// let myChart = new Chart(ctx, {
//   type: "line",
//   data: {
//     datasets: [
//       {
//         label: "Low",
//         backgroundColor: colors[0],
//         borderColor: colors[0],
//         cubicInterpolationMode: "monotone",
//         tension: 0.4,
//         data: dataLow,
//       },
//       {
//         label: "High",
//         backgroundColor: colors[1],
//         borderColor: colors[1],
//         cubicInterpolationMode: "monotone",
//         tension: 0.4,
//         data: dataHigh,
//       },
//     ],
//   },
//   options: {
//     scales: {
//       x: {
//         type: "time",
//         min: 1655403669818 - 1000 * 60,
//         max: 1655407200000 + 1000 * 60,
//       },
//     },
//   },
// });
