import React, { useEffect, useRef, useState, forwardRef, useImperativeHandle } from 'react';
import { Line } from 'react-chartjs-2';
import 'chart.js/auto';
import { getMetricsAll } from '../services/api';

const palette = {
  fcfs: { wt: 'rgba(37, 99, 235, 1)', tat: 'rgba(37, 99, 235, 0.6)' },
  ljf:  { wt: 'rgba(16, 185, 129, 1)', tat: 'rgba(16, 185, 129, 0.6)' },
  lifo: { wt: 'rgba(249, 115, 22, 1)', tat: 'rgba(249, 115, 22, 0.6)' },
};

const MetricsGraph = forwardRef((props, ref) => {
  const [results, setResults] = useState([]);
  const [summary, setSummary] = useState([]);
  const didInit = useRef(false);

  const fetchMetrics = async () => {
    try {
      const data = await getMetricsAll();
      const arr = Array.isArray(data.results) ? data.results : [];
      const order = { fcfs: 0, ljf: 1, lifo: 2 };
      arr.sort((a, b) => (order[a.algo] ?? 9) - (order[b.algo] ?? 9));
      setResults(arr);
      setSummary(
        arr.map(r => ({
          algo: r.algo,
          avg_waiting: r.avg_waiting ?? 0,
          avg_tat: r.avg_tat ?? 0,
          throughput: r.throughput ?? 0,
        }))
      );
    } catch (e) {
      console.error('Failed to fetch metrics:', e);
    }
  };

  useImperativeHandle(ref, () => ({ fetchMetrics }), []);

  useEffect(() => {
    if (didInit.current) return;
    didInit.current = true;
    fetchMetrics();
  }, []);

  const toChart = (algoKey) => {
    const r = results.find(x => x.algo === algoKey);
    const metrics = r?.metrics || [];
    return {
      labels: metrics.map(m => `Task ${m.task_id}`),
      datasets: [
        { label: 'Waiting Time',   data: metrics.map(m => Math.round(m.waiting_time)),   borderColor: palette[algoKey].wt,  fill: false },
        { label: 'Turnaround Time',data: metrics.map(m => Math.round(m.turnaround_time)), borderColor: palette[algoKey].tat, fill: false },
      ],
    };
  };

  const chartOptions = (title) => ({
    responsive: true,
    plugins: {
      title: { display: true, text: title },
    },
    scales: {
      x: {
        title: {
          display: true,
          text: 'Tasks',
          font: { size: 14 }
        }
      },
      y: {
        title: {
          display: true,
          text: 'Time (s)',
          font: { size: 14 }
        }
      }
    }
  });

  return (
    <div style={{ marginTop: 30 }}>
      <h2>Performance Metrics (FCFS vs LJF vs LIFO)</h2>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 32 }}>
        <section>
          <h3>FCFS</h3>
          <Line data={toChart('fcfs')} options={chartOptions('FCFS Task Metrics')} />
        </section>

        <section>
          <h3>LJF</h3>
          <Line data={toChart('ljf')} options={chartOptions('LJF Task Metrics')} />
        </section>

        <section>
          <h3>LIFO</h3>
          <Line data={toChart('lifo')} options={chartOptions('LIFO Task Metrics')} />
        </section>
      </div>

      <div style={{ marginTop: 16 }}>
        {summary.map(s => (
          <p key={s.algo}>
            <strong>{s.algo.toUpperCase()}</strong> — <strong>Avg WT:</strong> {s.avg_waiting}{' '}
            <strong>Avg TAT:</strong> {s.avg_tat}{' '}
            <strong>Throughput:</strong> {s.throughput}
          </p>
        ))}
      </div>
    </div>
  );
});

export default MetricsGraph;