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

async function update(chart) {
  const res = await fetch("/api");
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
  chart.options.scales = getScales();
  chart.update();
}

async function main() {
  const chart = createChart();
  await update(chart).catch(console.warn);
  setInterval(() => {
    update(chart).catch(console.warn);
  }, 5000);
}

main();
